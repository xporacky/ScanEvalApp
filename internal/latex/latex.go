package latex

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/gorm"

	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/logging"
	"log/slog"
)

// CompileLatexToPDF compiles a LaTeX template into a PDF.
// It creates a temporary .tex file, writes the LaTeX content to it,
// compiles the LaTeX file using pdflatex, and returns the generated PDF as bytes.
// Returns an error if any step fails.
func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Create a temporary file to store the LaTeX content
	texFile, err := os.CreateTemp(common.TEMPORARY_PDF_PATH, "*.tex")
	if err != nil {
		errorLogger.Error("Failed to create temporary LaTeX file", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	defer func() {
		// Kontrola chyby pri mazani temp suboru
		if err := files.DeleteFile(texFile.Name()); err != nil {
			errorLogger.Error("Error deleting temporary LaTeX file", slog.Group("CRITICAL", slog.String("error", err.Error())))
		}
	}()

	// Write LaTeX content into the file
	if _, err = texFile.Write(latexContent); err != nil {
		errorLogger.Error("Error writing to .tex file", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	// Set the output directory for the PDF file
	if err := texFile.Close(); err != nil {
		errorLogger.Error("Error closing temporary LaTeX file", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	logger.Info("LaTeX file created", slog.String("file_path", texFile.Name()))

	outputDir, err := os.MkdirTemp(common.TEMPORARY_PDF_PATH, "latex_output")
	if err != nil {
		errorLogger.Error("Failed to create temporary output directory", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}
	defer os.RemoveAll(outputDir)

	// Run pdflatex command to compile the LaTeX file
	cmd := exec.Command("pdflatex", "-output-directory", outputDir, texFile.Name())
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err = cmd.Run(); err != nil {
		errorLogger.Error("Error compiling LateX to PDF", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		errorLogger.Error("Error compiling LaTeX to PDF", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	logger.Info("PDF compiled")
	return pdfBytes, nil
}

// ReplaceTemplatePlaceholders replaces placeholders in a LaTeX template with the provided data.
// It parses the LaTeX template, executes it with the provided data, and returns the updated LaTeX content.
// Returns an error if parsing or template execution fails.
func ReplaceTemplatePlaceholders(templateContent []byte, data TemplateData) ([]byte, error) {
	errorLogger := logging.GetErrorLogger()

	// Parse the LaTeX template content
	tmpl, err := template.New("latex").Parse(string(templateContent))
	if err != nil {
		errorLogger.Error("Error parsing LaTeX template", slog.String("error", err.Error()))
		return nil, err
	}

	var output bytes.Buffer
	// Apply the data to the template
	err = tmpl.Execute(&output, data)
	if err != nil {
		errorLogger.Error("Error replacing placeholders in template", slog.String("error", err.Error()))
		return nil, err
	}

	return output.Bytes(), nil
}

// MergePDFs merges two PDFs into a single PDF using the pdfunite utility.
// The PDFs are combined in the order provided, and the output is written to the specified path.
// Returns an error if the merging process fails.
func MergePDFs(pdf1Path, pdf2Path, outputPath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Execute the pdfunite command to merge PDFs
	cmd := exec.Command("pdfunite", pdf1Path, pdf2Path, outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	logger.Debug("Merging PDFs", slog.String("pdf1", pdf1Path), slog.String("pdf2", pdf2Path))

	if err := cmd.Run(); err != nil {
		errorLogger.Error("Error merging PDFs", slog.String("error", err.Error()))
		return err
	}
	logger.Info("PDFs merged", slog.String("output_path", outputPath))
	return nil
}

// ParallelGeneratePDFs generates PDFs for students in parallel by processing each student in separate goroutines.
// It fetches unique rooms, retrieves student data from the database,
// and generates a LaTeX based PDF for each student by replacing template placeholders with student data.
// The generated PDFs are merged into a single PDF, which is saved to the specified output path.
// Returns an error and the path of the final merged PDF.
func ParallelGeneratePDFs(db *gorm.DB, examID uint, templatePath) (string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	var rooms []string
	// Fetch unique room names from the database
	if err := db.Model(&models.Student{}).Where("exam_id = ?", examID).Distinct().Pluck("room", &rooms).Error; err != nil {
		errorLogger.Error("Error fetching distinct rooms", slog.String("error", err.Error()))
		return "", err
	}

	// Synchronize goroutines using a WaitGroup
	var wg sync.WaitGroup
	var pdfMergeMutex sync.Mutex
	var processedCount int64

	// Variable to store the path of the main PDF
	var mainPDFPath string
	var mainPDFSet bool

	// nacitanie testu
	var exam models.Exam
	if err := db.First(&exam, examID).Error; err != nil {
		errorLogger.Error("Error fetching exam details", "exam_id", examID, slog.String("error", err.Error()))
		return "", err
	}

	logger.Debug("Starting parallel PDF generation")

	// Measure the total time for PDF generation and merging
	startTime := time.Now()

	// Loop through each room and process students due to correct ordering of PDFs
	for _, room := range rooms {
		logger.Info("Processing students in room", slog.String("room", room))

		// Fetch all students in the current room
		var students []models.Student
		if err := db.Where("room = ? AND exam_id = ?", room, examID).Find(&students).Error; err != nil {
			errorLogger.Error("Error fetching students", slog.String("error", err.Error()))
			return "", err
		}
		// Generate PDFs concurrently for each student
		for _, student := range students {
			wg.Add(1)
			go func(student models.Student) {
				defer wg.Done()

				studentStartTime := time.Now()

				// Load LaTeX template
				latexTemplate, err := os.ReadFile(templatePath)
				if err != nil {
					errorLogger.Error("Error reading LaTeX template for student", "student_id", student.ID, slog.String("error", err.Error()))
					return
				}

				// // Load the exam for the student
				// if err := db.First(&exam, student.ExamID).Error; err != nil {
				// 	errorLogger.Error("Error fetching exam for student", "student_id", student.ID, slog.String("error", err.Error()))
				// 	return
				// }

				// Prepare the data to replace placeholders in the LaTeX template
				data := TemplateData{
					ID:        fmt.Sprintf("%d", student.RegistrationNumber),
					Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
					Datum:     exam.Date.Format("02. 01. 2006"),
					Miestnost: student.Room,
					Cas:       exam.Date.Format("15:04"),
					Bloky:     exam.QuestionCount,
					QrCode:    fmt.Sprintf("%d", student.ID),
				}

				// Replace placeholders in the LaTeX template with the student data
				updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
				if err != nil {
					errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, slog.String("error", err.Error()))
					return
				}

				// Compile the LaTeX template into a PDF
				studentPDF, err := CompileLatexToPDF(updatedLatex)
				if err != nil {
					errorLogger.Error("Error generating PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
					return
				}

				// Save the generated PDF for the student
				studentPDFPath := filepath.Join(common.TEMPORARY_PDF_PATH, fmt.Sprintf("student_%d.pdf", student.ID))
				if err := os.WriteFile(studentPDFPath, studentPDF, common.FILE_PERMISSION); err != nil {
					errorLogger.Error("Error saving PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
					return
				}

				// Lock the PDF merge operation to ensure it happens sequentially
				pdfMergeMutex.Lock()
				defer pdfMergeMutex.Unlock()

				// Set the first student PDF as the main PDF
				if !mainPDFSet {
					mainPDFPath = studentPDFPath
					mainPDFSet = true
					logger.Info("Set initial main PDF for student", "student_id", student.ID)
				} else {
					// Merge the new student PDF with the existing main PDF
					mergedPDFPath := filepath.Join(common.TEMPORARY_PDF_PATH, "merged.pdf")
					if err := MergePDFs(mainPDFPath, studentPDFPath, mergedPDFPath); err != nil {
						errorLogger.Error("Error merging PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
						return
					}

					// Rename the merged PDF to be the main PDF
					if err := os.Rename(mergedPDFPath, mainPDFPath); err != nil {
						errorLogger.Error("Error updating main PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
						return
					}

					// Delete the temp student pdf
					defer func() {
						if err := files.DeleteFile(studentPDFPath); err != nil {
							errorLogger.Error("Error removing temporary PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
						}
					}()

				}

				// Increment the processed PDF count and log it
				processedCount++
				studentDuration := time.Since(studentStartTime)
				logger.Debug("Generovanie PDF",
					"spracovaných", processedCount,
					"celkovo", len(students),
					"exam", exam.Title,
					"id študenta", student.ID,
					"dokončené za", studentDuration)
			}(student)
		}

		// Waiting for all goroutines
		wg.Wait()
	}

	// TODO -> podla zmeny DB mozno nejak inak premenovat vysledny subor, napr zobrat nieco z roku alebo neviem...
	// Create a final pdf
	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return "", err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return "", err
	}

	safeTitle := common.SanitizeFilename(exam.Title)
	finalPDFPath := filepath.Join(absDirPath, fmt.Sprintf("%s%d.pdf", safeTitle, exam.ID))

	srcFile, err := os.Open(mainPDFPath)
	if err != nil {
		errorLogger.Error("error opening source PDF", slog.String("error", err.Error()))
		return "", err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(finalPDFPath)
	if err != nil {
		errorLogger.Error("error creating destination PDF", slog.String("error", err.Error()))
		return "", err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		errorLogger.Error("error copying PDF", slog.String("error", err.Error()))
		return "", err
	}

	// Files will be closed automatically via defer statements
	if err := os.Remove(mainPDFPath); err != nil {
		errorLogger.Error("error removing original PDF", slog.String("error", err.Error()))
		return "", err
	}
	// Measure the total time
	duration := time.Since(startTime)
	logger.Debug("Celkový čas generovania PDF", "duration", duration)

	absFinalPDFPath, err := filepath.Abs(finalPDFPath)
	if err != nil {
		errorLogger.Error("Nepodarilo sa získať absolútnu cestu k výslednému PDF", slog.String("error", err.Error()))
		return "", err
	}
	logger.Info("Výsledné PDF úspešne uložené do", slog.String("output_PDF_path", absFinalPDFPath))

	return absFinalPDFPath, nil
}

// PrintSheet generates a PDF for a student based on their registration number.
// It fetches student and exam data, prepares the LaTeX content, replaces placeholders,
// compiles the LaTeX into a PDF, and then saves the resulting PDF to a file.
// Returns an error if something goes wrong, and the file path of the generated PDF.
func PrintSheet(db *gorm.DB, registrationNumber int) (string, error) {

	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Find student by registrationNumber
	student, err := FindStudentByRegistrationNumber(db, registrationNumber)
	if err != nil {
		errorLogger.Error("Error finding student", "registration_number", registrationNumber, slog.String("error", err.Error()))
		return "", err
	}

	logger.Info("Student found", "student_id", student.ID, "registration_number", student.RegistrationNumber)

	// Load the latex template
	latexTemplate, err := os.ReadFile(common.TEMPLATE_PATH)
	if err != nil {
		errorLogger.Error("Error reading LaTeX template for student", "student_id", student.ID, slog.String("error", err.Error()))
		return "", err
	}

	logger.Info("LaTeX template loaded", "template_path", common.TEMPLATE_PATH)

	// Load the exam for student from the database
	var exam models.Exam
	if err := db.First(&exam, student.ExamID).Error; err != nil {
		errorLogger.Error("Error fetching exam for student", "student_id", student.ID, slog.String("error", err.Error()))
		return "", err
	}
	logger.Info("Exam fetched for student", "exam_id", exam.ID, "exam_title", exam.Title)

	// Create data template based on student's data from the database
	data := TemplateData{
		ID:        fmt.Sprintf("%d", student.RegistrationNumber),
		Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
		Datum:     exam.Date.Format("02. 01. 2006"), // datum v tvare DD. MM. YYYY
		Miestnost: student.Room,
		Cas:       exam.Date.Format("15:04"), // čas v tvare HH:MM
		Bloky:     exam.QuestionCount,
		QrCode:    fmt.Sprintf("%d", student.ID),
	}

	logger.Info("Template data prepared", "student_id", student.ID, "registration_number", student.RegistrationNumber)

	// Replace placeholder by student's data
	updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
	if err != nil {
		errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, slog.String("error", err.Error()))
		return "", err
	}
	logger.Info("Placeholders replaced successfully for student", "student_id", student.ID)

	// Compile the pdf
	studentPDF, err := CompileLatexToPDF(updatedLatex)
	if err != nil {
		errorLogger.Error("Error generating PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
		return "", err
	}
	logger.Info("PDF generated for student", "student_id", student.ID)

	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return "", err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return "", err
	}
	// Save the student's compiled PDF
	studentPDFPath := filepath.Join(absDirPath, fmt.Sprintf("student_%d.pdf", student.RegistrationNumber))
	if err := os.WriteFile(studentPDFPath, studentPDF, common.FILE_PERMISSION); err != nil {
		errorLogger.Error("Error saving PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
		return "", err
	}

	if err != nil {
		errorLogger.Error("Chyba pri získavaní absolútnej cesty k PDF", "student_id", student.ID, slog.String("error", err.Error()))
	} else {
		logger.Info("PDF saved successfully for student",
			"student_id", student.ID,
			"pdf_path", studentPDFPath,
		)
	}

	return studentPDFPath, nil
}
