package latex

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/gorm"

	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/logging"
	"log/slog"
)

// funkcia na kompilaciu LaTeX sablony do PDF
func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	texFile, err := os.CreateTemp(TemporaryPDFPath, "*.tex")
	if err != nil {
		errorLogger.Error("Failed to create temporary LaTeX file", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}
	defer files.DeleteFile(texFile.Name())

	if _, err = texFile.Write(latexContent); err != nil {
		errorLogger.Error("Error writing to .tex file", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}
	texFile.Close()

	logger.Info("LaTeX file created", slog.String("file_path", texFile.Name()))

	outputDir, err := os.MkdirTemp(TemporaryPDFPath, "latex_output")
	if err != nil {
		errorLogger.Error("Failed to create temporary output directory", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}
	defer os.RemoveAll(outputDir)

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

// funkcia na nahradenie placeholderov v LaTeX sablone
func ReplaceTemplatePlaceholders(templateContent []byte, data TemplateData) ([]byte, error) {
	errorLogger := logging.GetErrorLogger()

	tmpl, err := template.New("latex").Parse(string(templateContent))
	if err != nil {
		errorLogger.Error("Error parsing LaTeX template", slog.String("error", err.Error()))
		return nil, err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		errorLogger.Error("Error replacing placeholders in template", slog.String("error", err.Error()))
		return nil, err
	}

	return output.Bytes(), nil
}

// mergovanie 2 pdf pomocou pdfunite kniznice
func MergePDFs(pdf1Path, pdf2Path, outputPath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

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

// funkcia na paralelne generovanie pdf
func ParallelGeneratePDFs(db *gorm.DB, templatePath, outputPDFPath string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// nacitanie vsetkych studentov z databazy
	var students []models.Student
	if err := db.Find(&students).Error; err != nil {
		errorLogger.Error("Error fetching students", slog.String("error", err.Error()))
		return err
	}

	// meranie celkoveho casu generovania a mergovania pdf
	startTime := time.Now()

	// synchronizacia prace s goroutines pomocou WaitGroup
	var wg sync.WaitGroup
	var pdfMergeMutex sync.Mutex
	var processedCount int64

	// Premenná na udržanie cesty k hlavnému PDF
	var mainPDFPath string
	var mainPDFSet bool // označuje, či bolo už hlavné PDF nastavené

	logger.Debug("Starting parallel PDF generation")

	// paralelne generovanie pdf pre studentov
	for _, student := range students {
		wg.Add(1)
		go func(student models.Student) {
			defer wg.Done()

			// meranie casu spracovania generovanie pdf pre studenta
			studentStartTime := time.Now()

			// Nacitanie LaTeX sablony
			latexTemplate, err := os.ReadFile(templatePath)
			if err != nil {
				errorLogger.Error("Error reading LaTeX template for student", "student_id", student.ID, slog.String("error", err.Error()))
				return
			}

			// nacitanie testu pre studenta z databazy
			var test models.Exam
			if err := db.First(&test, student.ExamID).Error; err != nil {
				errorLogger.Error("Error fetching test for student", "student_id", student.ID, slog.String("error", err.Error()))
				return
			}

			// vytvorenie dat, ktore budu nacitane namiesto placeholderov v LaTeX sablone
			data := TemplateData{
				ID:        fmt.Sprintf("%d", student.RegistrationNumber),
				Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
				Datum:     test.Date.Format("02. 01. 2006"), // datum v tvare DD. MM. YYYY
				Miestnost: student.Room,
				Cas:       test.Date.Format("15:04"), // čas v tvare HH:MM
				Bloky:     test.QuestionCount,
				QrCode:    fmt.Sprintf("%d", student.ID),
			}

			// nahradenie placeholderov v LaTeX sablone udajmi studenta
			updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
			if err != nil {
				errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, slog.String("error", err.Error()))
				return
			}

			// kompilacia latex sablony do pdf studenta
			studentPDF, err := CompileLatexToPDF(updatedLatex)
			if err != nil {
				errorLogger.Error("Error generating PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
				return
			}

			// ulozenie generovaneho/kompilovaneho pdf studenta do docasneho suboru
			studentPDFPath := filepath.Join(TemporaryPDFPath, fmt.Sprintf("student_%d.pdf", student.ID))
			if err := os.WriteFile(studentPDFPath, studentPDF, FilePermission); err != nil {
				errorLogger.Error("Error saving PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
				return
			}

			pdfMergeMutex.Lock()
			defer pdfMergeMutex.Unlock()

			// Ak ešte nie je nastavené hlavné PDF, použijeme toto ako hlavné
			if !mainPDFSet {
				mainPDFPath = studentPDFPath
				mainPDFSet = true
				errorLogger.Error("Set initial main PDF for student", "student_id", student.ID)
			} else {
				// zlucenie generovaneho pdf so zakladnym pdf
				mergedPDFPath := filepath.Join(TemporaryPDFPath, "merged.pdf")
				if err := MergePDFs(mainPDFPath, studentPDFPath, mergedPDFPath); err != nil {
					errorLogger.Error("Error merging PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
					return
				}

				// nahradenie hlavneho pdf novym zlucenym pdf
				if err := os.Rename(mergedPDFPath, mainPDFPath); err != nil {
					errorLogger.Error("Error updating main PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
					return
				}

				// Odstranenie docasneho pdf studenta s defer, ktore sa vykona vzdy na konci funkcie
				defer func() {
					if err := files.DeleteFile(studentPDFPath); err != nil {
						errorLogger.Error("Error removing temporary PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
					}
				}()

			}

			// Zvýšenie počtu spracovaných PDF a výpis stavu
			processedCount++
			studentDuration := time.Since(studentStartTime)
			logger.Debug("Generovanie PDF",
				"spracovaných", processedCount,
				"celkovo", len(students),
				"test", test.Title,
				"id študenta", student.ID,
				"dokončené za", studentDuration)
		}(student)
	}

	// cakanie na vsetky goroutines
	wg.Wait()

	// presun hlavného PDF na finálnu cestu
	if err := os.Rename(mainPDFPath, outputPDFPath); err != nil {
		errorLogger.Error("error moving final PDF", slog.String("error", err.Error()))
		return err
	}

	// zmeranie celkoveho casu behu programu
	duration := time.Since(startTime)
	logger.Debug("Celkový čas generovania PDF", "duration", duration)

	logger.Info("Výsledné PDF úspešne uložené do", slog.String("output_PDF_path", outputPDFPath))
	return nil
}

// FindStudentByRegistrationNumber nájde študenta v DB podľa RegistrationNumber
func FindStudentByRegistrationNumber(db *gorm.DB, registrationNumber int) (*models.Student, error) {
	var student models.Student
	if err := db.Where("registration_number = ?", registrationNumber).First(&student).Error; err != nil {
		return nil, fmt.Errorf("student not found with RegistrationNumber %d: %w", registrationNumber, err)
	}
	return &student, nil
}

func PrintSheet(db *gorm.DB, registrationNumber int) error {

	// Inicializácia loggera
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Najprv nájdi študenta podľa RegistrationNumber
	student, err := FindStudentByRegistrationNumber(db, registrationNumber)
	if err != nil {
		errorLogger.Error("Error finding student", "registration_number", registrationNumber, slog.String("error", err.Error()))
		return err // Ak študent neexistuje, vráti sa chyba
	}

	// Logovanie úspešného nájdenia študenta
	logger.Info("Student found", "student_id", student.ID, "registration_number", student.RegistrationNumber)

	// Nacitanie LaTeX sablony
	latexTemplate, err := os.ReadFile(TemplatePath)
	if err != nil {
		errorLogger.Error("Error reading LaTeX template for student", "student_id", student.ID, slog.String("error", err.Error()))
		return err
	}
	// Logovanie načítania LaTeX šablóny
	logger.Info("LaTeX template loaded", "template_path", TemplatePath)

	// Nacitanie testu pre studenta z databazy
	var test models.Exam
	if err := db.First(&test, student.ExamID).Error; err != nil {
		errorLogger.Error("Error fetching test for student", "student_id", student.ID, slog.String("error", err.Error()))
		return err
	}
	// Logovanie úspešného načítania testu
	logger.Info("Test fetched for student", "test_id", test.ID, "test_title", test.Title)

	// Vytvorenie dat, ktore budu nacitane namiesto placeholderov v LaTeX sablone
	data := TemplateData{
		ID:        fmt.Sprintf("%d", student.RegistrationNumber),
		Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
		Datum:     test.Date.Format("02. 01. 2006"), // datum v tvare DD. MM. YYYY
		Miestnost: student.Room,
		Cas:       test.Date.Format("15:04"), // čas v tvare HH:MM
		Bloky:     test.QuestionCount,
		QrCode:    fmt.Sprintf("%d", student.ID),
	}
	// Logovanie dát, ktoré sa použijú pre šablónu
	logger.Info("Template data prepared", "student_id", student.ID, "registration_number", student.RegistrationNumber)

	// Nahradenie placeholderov v LaTeX sablone udajmi studenta
	updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
	if err != nil {
		errorLogger.Error("Error replacing placeholders for student", "student_id", student.ID, slog.String("error", err.Error()))
		return err
	}
	// Logovanie úspešného nahradenia placeholderov
	logger.Info("Placeholders replaced successfully for student", "student_id", student.ID)

	// Kompilacia latex sablony do pdf studenta
	studentPDF, err := CompileLatexToPDF(updatedLatex)
	if err != nil {
		errorLogger.Error("Error generating PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
		return err
	}
	// Logovanie úspešnej kompilácie PDF
	logger.Info("PDF generated for student", "student_id", student.ID)

	// Ulozenie generovaneho/kompilovaneho pdf studenta do docasneho suboru
	studentPDFPath := filepath.Join(TemporaryPDFPath, fmt.Sprintf("student_%d.pdf", student.RegistrationNumber))
	if err := os.WriteFile(studentPDFPath, studentPDF, FilePermission); err != nil {
		errorLogger.Error("Error saving PDF for student", "student_id", student.ID, slog.String("error", err.Error()))
		return err
	}
	// Logovanie úspešného uloženia PDF
	logger.Info("PDF saved successfully for student", "student_id", student.ID, "pdf_path", studentPDFPath)

	return nil
}
