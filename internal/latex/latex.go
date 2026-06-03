package latex

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"text/template"
	"time"

	"gorm.io/gorm"

	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/logging"
	"ScanEvalApp/internal/scanprocessing"
)

func buildIDDigits(registrationNumber int) []string {
	id := fmt.Sprintf("%07d", registrationNumber)
	if len(id) > 7 {
		id = id[:7]
	}
	digits := make([]string, 7)
	for i, r := range id {
		if i >= len(digits) {
			break
		}
		digits[i] = string(r)
	}
	return digits
}

func buildQRCodePayload(student models.Student, exam models.Exam) string {
	datum := exam.Date
	if !student.ExamDate.IsZero() {
		datum = student.ExamDate
	}
	return latexEscape(fmt.Sprintf(
		"ID:%07d | MENO:%s %s | DATUM:%s | MIESTNOST:%s",
		student.RegistrationNumber,
		student.Name,
		student.Surname,
		datum.Format("02.01.2006"),
		student.Room,
	))
}

func buildTemplateData(student models.Student, exam models.Exam) TemplateData {
	datum := exam.Date.Format("02. 01. 2006")
	cas := exam.Date.Format("15:04")
	if !student.ExamDate.IsZero() {
		datum = student.ExamDate.Format("02. 01. 2006")
	}
	if student.ExamTime != "" {
		cas = student.ExamTime
	}
	return TemplateData{
		ID:        latexEscape(fmt.Sprintf("%d", student.RegistrationNumber)),
		IDDigits:  buildIDDigits(student.RegistrationNumber),
		Meno:      latexEscape(fmt.Sprintf("%s %s", student.Name, student.Surname)),
		ShowName:  exam.ShowName,
		Datum:     latexEscape(datum),
		Miestnost: latexEscape(student.Room),
		Cas:       latexEscape(cas),
		Bloky:     exam.QuestionCount,
		QrCode:    buildQRCodePayload(student, exam),
		TestName:  latexEscape(exam.Title),
	}
}

func formatLatexCompileError(stdout, stderr string) string {
	if stderr != "" {
		return stderr
	}

	if stdout != "" {
		return stdout
	}

	return "pdflatex failed without output"
}

// CompileLatexToPDF compiles a LaTeX template into a PDF.
func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Create a temporary file to store the LaTeX content
	texFile, err := os.CreateTemp(common.TEMPORARY_PDF_PATH, "*.tex")
	if err != nil {
		errorLogger.Error("Failed to create temporary LaTeX file", "error", err.Error())
		return nil, err
	}

	defer func() {
		if err := files.DeleteFile(texFile.Name()); err != nil {
			errorLogger.Error("Error deleting temporary LaTeX file", "error", err.Error())
		}
	}()

	// Write LaTeX content into the file
	if _, err = texFile.Write(latexContent); err != nil {
		errorLogger.Error("Error writing to .tex file", "error", err.Error())
		return nil, err
	}

	// Set the output directory for the PDF file
	if err := texFile.Close(); err != nil {
		errorLogger.Error("Error closing temporary LaTeX file", "error", err.Error())
		return nil, err
	}

	logger.Info("LaTeX file created", "file_path", texFile.Name())

	outputDir, err := os.MkdirTemp(common.TEMPORARY_PDF_PATH, "latex_output")
	if err != nil {
		errorLogger.Error("Failed to create temporary output directory", "error", err.Error())
		return nil, err
	}
	defer os.RemoveAll(outputDir)

	// DEBUG: Save the LaTeX content for manual inspection
	debugTexPath := filepath.Join(outputDir, "debug_latex.tex")
	if err := os.WriteFile(debugTexPath, latexContent, 0644); err == nil {
		logger.Debug("LaTeX content saved for debugging", "debug_file", debugTexPath)
	}

	// Run pdflatex command to compile the LaTeX file - CAPTURE BOTH STDOUT AND STDERR
	cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error", "-output-directory", outputDir, texFile.Name())

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Debug("Running pdflatex command", "command", cmd.String())

	if err = cmd.Run(); err != nil {
		latexStdout := stdout.String()
		latexStderr := stderr.String()
		combinedOutput := formatLatexCompileError(latexStdout, latexStderr)

		errorLogger.Error("=== LaTeX COMPILATION FAILED ===",
			"error", err.Error(),
			"stdout", latexStdout,
			"stderr", latexStderr,
			"tex_file", texFile.Name(),
			"output_dir", outputDir)

		// Check if output files were created
		pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
		if _, pdfErr := os.Stat(pdfPath); pdfErr == nil {
			errorLogger.Error("PDF file was created despite error", "pdf_path", pdfPath)
		} else {
			errorLogger.Error("PDF file was NOT created", "expected_pdf_path", pdfPath)
		}

		// List files in output directory
		if files, err := os.ReadDir(outputDir); err == nil {
			errorLogger.Error("Files in output directory:")
			for _, file := range files {
				errorLogger.Error("Output file:", "file", file.Name(), "is_dir", file.IsDir())
			}
		}

		return nil, fmt.Errorf("latex compilation failed: %s", combinedOutput)
	}

	pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		errorLogger.Error("Error reading compiled PDF",
			"error", err.Error(),
			"pdf_path", pdfPath)

		// List files to see what was generated
		if files, listErr := os.ReadDir(outputDir); listErr == nil {
			errorLogger.Error("Files in output directory after compilation:")
			for _, file := range files {
				errorLogger.Error("Generated file:", "file", file.Name(), "is_dir", file.IsDir())
			}
		}
		return nil, err
	}

	logger.Info("PDF compiled successfully", "pdf_size_bytes", len(pdfBytes))
	return pdfBytes, nil
}

// ReplaceTemplatePlaceholders replaces placeholders in a LaTeX template with the provided data.
func ReplaceTemplatePlaceholders(templateContent []byte, data TemplateData) ([]byte, error) {
	errorLogger := logging.GetErrorLogger()

	// Parse the LaTeX template content
	tmpl, err := template.New("latex").Parse(string(templateContent))
	if err != nil {
		errorLogger.Error("Error parsing LaTeX template", "error", err.Error())
		return nil, err
	}

	var output bytes.Buffer
	// Apply the data to the template
	err = tmpl.Execute(&output, data)
	if err != nil {
		errorLogger.Error("Error replacing placeholders in template", "error", err.Error())
		return nil, err
	}

	return output.Bytes(), nil
}

// MergePDFs merges two PDFs into a single PDF using the pdfunite utility.
func MergePDFs(pdf1Path, pdf2Path, outputPath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Execute the pdfunite command to merge PDFs
	cmd := exec.Command("pdfunite", pdf1Path, pdf2Path, outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	logger.Debug("Merging PDFs", "pdf1", pdf1Path, "pdf2", pdf2Path)

	if err := cmd.Run(); err != nil {
		errorLogger.Error("Error merging PDFs", "error", err.Error())
		return err
	}
	logger.Info("PDFs merged", "output_path", outputPath)
	return nil
}

// generateBlankPagePDF generates a single blank A4 page PDF and returns its file path.
func generateBlankPagePDF() (string, error) {
	blankLatex := []byte(`\documentclass{article}
\usepackage[a4paper,margin=1in]{geometry}
\begin{document}
\null
\end{document}`)
	pdfBytes, err := CompileLatexToPDF(blankLatex)
	if err != nil {
		return "", fmt.Errorf("failed to compile blank page PDF: %w", err)
	}
	blankPath := filepath.Join(common.TEMPORARY_PDF_PATH, "blank_page.pdf")
	if err := os.WriteFile(blankPath, pdfBytes, common.FILE_PERMISSION); err != nil {
		return "", fmt.Errorf("failed to write blank page PDF: %w", err)
	}
	return blankPath, nil
}

// generateStudentPDFs generates PDFs for the given students in parallel, merges them, and saves to outputPath.
func generateStudentPDFs(students []models.Student, exam models.Exam, templatePath string, outputPath string) error {
	errorLogger := logging.GetErrorLogger()
	logger := logging.GetLogger()

	latexTemplate, err := os.ReadFile(templatePath)
	if err != nil {
		errorLogger.Error("Error reading LaTeX template", "error", err.Error(), "template_path", templatePath)
		return err
	}

	type studentPDFResult struct {
		index   int
		pdfPath string
		id      uint
		err     error
	}

	results := make([]studentPDFResult, len(students))
	var wg sync.WaitGroup
	var genErrMu sync.Mutex
	var genErr error

	// If pages per student is odd, each student's PDF needs a trailing blank page
	// so that double-sided printing does not mix pages from different students.
	// pages per student = ceil(QuestionCount / 20)
	pagesPerStudent := (exam.QuestionCount + scanprocessing.NUMBER_OF_QUESTIONS_PER_PAGE - 1) / scanprocessing.NUMBER_OF_QUESTIONS_PER_PAGE
	needsBlankPage := pagesPerStudent > 1 && pagesPerStudent%2 != 0

	var blankPDFPath string
	if needsBlankPage {
		var blankErr error
		blankPDFPath, blankErr = generateBlankPagePDF()
		if blankErr != nil {
			errorLogger.Error("Failed to generate blank page PDF", "error", blankErr.Error())
			return fmt.Errorf("failed to generate blank page PDF: %w", blankErr)
		}
		defer os.Remove(blankPDFPath)
		logger.Info("Blank page PDF prepared for odd page count", "pages_per_student", pagesPerStudent)
	}

	for i, student := range students {
		wg.Add(1)
		go func(idx int, s models.Student) {
			defer wg.Done()
			data := buildTemplateData(s, exam)
			updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
			if err != nil {
				genErrMu.Lock()
				if genErr == nil {
					genErr = err
				}
				genErrMu.Unlock()
				return
			}
			studentPDF, err := CompileLatexToPDF(updatedLatex)
			if err != nil {
				genErrMu.Lock()
				if genErr == nil {
					genErr = err
				}
				genErrMu.Unlock()
				return
			}
			pdfPath := filepath.Join(common.TEMPORARY_PDF_PATH, fmt.Sprintf("student_%d.pdf", s.ID))
			if err := os.WriteFile(pdfPath, studentPDF, common.FILE_PERMISSION); err != nil {
				genErrMu.Lock()
				if genErr == nil {
					genErr = err
				}
				genErrMu.Unlock()
				return
			}
			if needsBlankPage {
				withBlankPath := filepath.Join(common.TEMPORARY_PDF_PATH, fmt.Sprintf("student_%d_blank.pdf", s.ID))
				if err := MergePDFs(pdfPath, blankPDFPath, withBlankPath); err != nil {
					genErrMu.Lock()
					if genErr == nil {
						genErr = fmt.Errorf("failed to append blank page for student %d: %w", s.ID, err)
					}
					genErrMu.Unlock()
					return
				}
				if err := os.Rename(withBlankPath, pdfPath); err != nil {
					genErrMu.Lock()
					if genErr == nil {
						genErr = fmt.Errorf("failed to rename blank-appended PDF for student %d: %w", s.ID, err)
					}
					genErrMu.Unlock()
					return
				}
			}
			results[idx] = studentPDFResult{index: idx, pdfPath: pdfPath, id: s.ID}
			logger.Debug("PDF vygenerované", "student_id", s.ID)
		}(i, student)
	}
	wg.Wait()

	if genErr != nil {
		return genErr
	}

	var mainPDFPath string
	var mainPDFSet bool

	for _, result := range results {
		if result.pdfPath == "" {
			continue
		}
		if !mainPDFSet {
			mainPDFPath = result.pdfPath
			mainPDFSet = true
		} else {
			mergedPDFPath := filepath.Join(common.TEMPORARY_PDF_PATH, fmt.Sprintf("merged_%d.pdf", result.id))
			if err := MergePDFs(mainPDFPath, result.pdfPath, mergedPDFPath); err != nil {
				errorLogger.Error("Error merging PDF", "student_id", result.id, "error", err.Error())
				_ = files.DeleteFile(result.pdfPath)
				return err
			}
			if err := os.Rename(mergedPDFPath, mainPDFPath); err != nil {
				return err
			}
			if err := files.DeleteFile(result.pdfPath); err != nil {
				errorLogger.Error("Error removing temporary PDF", "student_id", result.id, "error", err.Error())
			}
		}
	}

	if !mainPDFSet {
		return fmt.Errorf("žiadne PDF pre skupinu")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), common.FILE_PERMISSION); err != nil {
		return err
	}

	srcFile, err := os.Open(mainPDFPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Remove(mainPDFPath)
}

// GenerateMultiDayPDFs generates one PDF per unique date+time+room combination.
// Output files are placed in {savePath}/pdf_output/ directory.
// Filename format: datum_cas_miestnost.pdf (e.g. 2026-04-20_08-00_A101.pdf)
func GenerateMultiDayPDFs(db *gorm.DB, examID uint) ([]string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	var exam models.Exam
	if err := db.First(&exam, examID).Error; err != nil {
		return nil, err
	}

	templatePath := common.GetTemplatePath(exam.OptionCount)
	if _, err := os.Stat(templatePath); err != nil {
		return nil, fmt.Errorf("template file does not exist: %s", templatePath)
	}

	var allStudents []models.Student
	if err := db.Where("exam_id = ?", examID).Find(&allStudents).Error; err != nil {
		return nil, err
	}

	// Zoskup podľa kombinácie dátum+čas+miestnosť
	groupMap := make(map[string][]models.Student)
	for _, s := range allStudents {
		key := fmt.Sprintf("%s_%s_%s", s.ExamDate.Format("2006-01-02"), s.ExamTime, s.Room)
		groupMap[key] = append(groupMap[key], s)
	}

	// Zoraď kľúče: dátum → čas → miestnosť
	groups := make([]string, 0, len(groupMap))
	for k := range groupMap {
		groups = append(groups, k)
	}
	sort.Strings(groups)

	dirPath, err := config.LoadLastPath()
	if err != nil {
		return nil, err
	}
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, err
	}
	printDir := filepath.Join(absDirPath, "pdf_output")
	if err := os.MkdirAll(printDir, 0755); err != nil {
		return nil, err
	}

	startTime := time.Now()
	var outputPaths []string

	for _, groupKey := range groups {
		students := groupMap[groupKey]

		// Zoraď vnútri skupiny podľa ID: prvá štvorica → celé ID
		sort.Slice(students, func(i, j int) bool {
			pi, pj := students[i].RegistrationNumber/1000, students[j].RegistrationNumber/1000
			if pi != pj {
				return pi < pj
			}
			return students[i].RegistrationNumber < students[j].RegistrationNumber
		})

		// Názov súboru: datum_cas_miestnost.pdf (nahraď ':' za '-' pre bezpečnosť)
		safeKey := common.SanitizeFilename(groupKey)
		outputPath := filepath.Join(printDir, fmt.Sprintf("%s.pdf", safeKey))

		if err := generateStudentPDFs(students, exam, templatePath, outputPath); err != nil {
			errorLogger.Error("Chyba pri generovaní PDF pre skupinu", "group", groupKey, "error", err.Error())
			return outputPaths, err
		}

		logger.Info("PDF pre skupinu vygenerované", "group", groupKey, "path", outputPath)
		outputPaths = append(outputPaths, outputPath)
	}

	logger.Info("Multi-day PDF generovanie dokončené", "groups", len(groups), "duration", time.Since(startTime))
	return outputPaths, nil
}

// ParallelGeneratePDFs generates PDFs for students in parallel by processing each student in separate goroutines.
func ParallelGeneratePDFs(db *gorm.DB, examID uint) (string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// DEBUG: Check current working directory
	if wd, err := os.Getwd(); err == nil {
		logger.Debug("Current working directory", "working_directory", wd)
	} else {
		errorLogger.Error("Error getting working directory", "error", err.Error())
	}

	logger.Debug("Starting PDF generation", "exam_id", examID)

	// Load exam details including option count
	var exam models.Exam
	if err := db.First(&exam, examID).Error; err != nil {
		errorLogger.Error("Error fetching exam details", "exam_id", examID, "error", err.Error())
		return "", err
	}

	// COMPREHENSIVE DEBUGGING - Check exam data
	logger.Debug("Exam details loaded",
		"exam_id", exam.ID,
		"exam_title", exam.Title,
		"option_count", exam.OptionCount,
		"question_count", exam.QuestionCount)

	// Check if OptionCount is valid
	if exam.OptionCount < common.MIN_OPTION_COUNT || exam.OptionCount > common.MAX_OPTION_COUNT {
		errorLogger.Error("Invalid option count in exam",
			"option_count", exam.OptionCount,
			"min_allowed", common.MIN_OPTION_COUNT,
			"max_allowed", common.MAX_OPTION_COUNT)
		return "", fmt.Errorf("invalid option count: %d", exam.OptionCount)
	}

	// Determine template path based on option count
	templatePath := common.GetTemplatePath(exam.OptionCount)

	// COMPREHENSIVE PATH DEBUGGING
	logger.Debug("=== TEMPLATE PATH DEBUGGING ===")
	logger.Debug("Template path result",
		"template_path", templatePath,
		"input_option_count", exam.OptionCount)

	// Check if template path is empty
	if templatePath == "" {
		errorLogger.Error("Template path is EMPTY - this is the problem!")
		return "", fmt.Errorf("template path is empty")
	}

	// Check if template file exists with detailed info
	fileInfo, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		errorLogger.Error("=== TEMPLATE FILE DOES NOT EXIST ===",
			"template_path", templatePath,
			"option_count", exam.OptionCount,
			"error", err.Error())

		// List available templates for debugging
		templateDir := filepath.Dir(templatePath)
		logger.Debug("Checking template directory", "directory", templateDir)

		// Check if directory exists
		if dirInfo, dirErr := os.Stat(templateDir); os.IsNotExist(dirErr) {
			errorLogger.Error("=== TEMPLATE DIRECTORY DOES NOT EXIST ===",
				"directory", templateDir,
				"error", dirErr.Error())

			// Show what directories do exist
			parentDir := filepath.Dir(templateDir)
			logger.Debug("Checking parent directory", "parent_directory", parentDir)
			if parentFiles, parentErr := os.ReadDir(parentDir); parentErr == nil {
				logger.Debug("Files in parent directory:")
				for _, file := range parentFiles {
					logger.Debug("Parent dir file:", "file", file.Name(), "is_dir", file.IsDir())
				}
			}
		} else if dirErr != nil {
			errorLogger.Error("Error checking template directory", "error", dirErr.Error())
		} else {
			logger.Debug("Template directory exists",
				"directory", templateDir,
				"is_dir", dirInfo.IsDir())
		}

		// List files in template directory
		files, err := os.ReadDir(templateDir)
		if err != nil {
			errorLogger.Error("Error reading template directory", "error", err.Error())
		} else {
			logger.Debug("=== AVAILABLE FILES IN TEMPLATE DIRECTORY ===")
			for _, file := range files {
				logger.Debug("Found file:",
					"file", file.Name(),
					"is_dir", file.IsDir(),
					"full_path", filepath.Join(templateDir, file.Name()))
			}

			// Check if any template files exist with similar names
			logger.Debug("=== LOOKING FOR SIMILAR TEMPLATE FILES ===")
			for i := common.MIN_OPTION_COUNT; i <= common.MAX_OPTION_COUNT; i++ {
				expectedFile := fmt.Sprintf("template_%d.tex", i)
				fullExpectedPath := filepath.Join(templateDir, expectedFile)
				if _, err := os.Stat(fullExpectedPath); err == nil {
					logger.Debug("Template file EXISTS",
						"file", expectedFile,
						"full_path", fullExpectedPath)
				} else {
					logger.Debug("Template file NOT FOUND",
						"file", expectedFile,
						"full_path", fullExpectedPath)
				}
			}
		}
		return "", fmt.Errorf("template file does not exist: %s", templatePath)
	} else if err != nil {
		errorLogger.Error("Error checking template file",
			"template_path", templatePath,
			"error", err.Error())
		return "", err
	} else {
		logger.Info("=== TEMPLATE FILE FOUND ===",
			"template_path", templatePath,
			"file_size", fileInfo.Size(),
			"file_mod_time", fileInfo.ModTime())
	}

	logger.Info("Using template", "template_path", templatePath, "option_count", exam.OptionCount)

	// Fetch all students and sort by last 3 digits of RegistrationNumber
	var allStudents []models.Student
	if err := db.Where("exam_id = ?", examID).Find(&allStudents).Error; err != nil {
		errorLogger.Error("Error fetching students", "error", err.Error())
		return "", err
	}
	sort.Slice(allStudents, func(i, j int) bool {
		return allStudents[i].RegistrationNumber%1000 < allStudents[j].RegistrationNumber%1000
	})

	startTime := time.Now()

	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", "error", err.Error())
		return "", err
	}
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", "error", err.Error())
		return "", err
	}

	safeTitle := common.SanitizeFilename(exam.Title)
	finalPDFPath := filepath.Join(absDirPath, fmt.Sprintf("%s%d.pdf", safeTitle, exam.ID))

	if err := generateStudentPDFs(allStudents, exam, templatePath, finalPDFPath); err != nil {
		return "", err
	}

	duration := time.Since(startTime)
	logger.Debug("Celkový čas generovania PDF", "duration", duration)

	absFinalPDFPath, err := filepath.Abs(finalPDFPath)
	if err != nil {
		errorLogger.Error("Nepodarilo sa získať absolútnu cestu k výslednému PDF", "error", err.Error())
		return "", err
	}
	logger.Info("Výsledné PDF úspešne uložené do", "output_PDF_path", absFinalPDFPath)

	return absFinalPDFPath, nil
}

// PrintSheet generates a PDF for a student based on their registration number.
func PrintSheet(db *gorm.DB, studentID uint) (string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Find student by primary key so registration numbers can repeat across exams.
	var student models.Student
	if err := db.First(&student, studentID).Error; err != nil {
		errorLogger.Error("Error finding student", "student_id", studentID, "error", err.Error())
		return "", err
	}

	logger.Info("Student found", "student_id", student.ID, "registration_number", student.RegistrationNumber)

	// Load the exam for student from the database
	var exam models.Exam
	if err := db.First(&exam, student.ExamID).Error; err != nil {
		errorLogger.Error("Error fetching exam for student", "student_id", student.ID, "error", err.Error())
		return "", err
	}
	logger.Info("Exam fetched for student", "exam_id", exam.ID, "exam_title", exam.Title)

	// Determine template path based on exam's option count
	templatePath := common.GetTemplatePath(exam.OptionCount)

	// Load the latex template with dynamic path
	latexTemplate, err := os.ReadFile(templatePath)
	if err != nil {
		errorLogger.Error("Error reading LaTeX template for student", "student_id", student.ID, "error", err.Error())
		return "", err
	}

	logger.Info("LaTeX template loaded", "template_path", templatePath, "option_count", exam.OptionCount)

	// Create data template based on student's data from the database
	data := buildTemplateData(student, exam)

	logger.Info("Template data prepared", "student_id", student.ID, "registration_number", student.RegistrationNumber)

	// Replace placeholder by student's data
	updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
	if err != nil {
		errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, "error", err.Error())
		return "", err
	}
	logger.Info("Placeholders replaced successfully for student", "student_id", student.ID)

	// Compile the pdf
	studentPDF, err := CompileLatexToPDF(updatedLatex)
	if err != nil {
		errorLogger.Error("Error generating PDF for student", "student_id", student.ID, "error", err.Error())
		return "", err
	}
	logger.Info("PDF generated for student", "student_id", student.ID)

	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", "error", err.Error())
		return "", err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", "error", err.Error())
		return "", err
	}

	// Save the student's compiled PDF
	studentPDFPath := filepath.Join(absDirPath, fmt.Sprintf("student_%d.pdf", student.RegistrationNumber))
	err = os.WriteFile(studentPDFPath, studentPDF, common.FILE_PERMISSION)
	if err != nil {
		errorLogger.Error("Chyba pri ukladaní PDF pre študenta",
			"student_id", student.ID,
			"error", err.Error())
		return "", err
	}

	logger.Info("PDF úspešne uložené pre študenta",
		"student_id", student.ID,
		"pdf_path", studentPDFPath,
	)

	return studentPDFPath, nil
}
