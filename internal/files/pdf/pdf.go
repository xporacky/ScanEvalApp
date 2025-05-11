package pdf

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/latex"
	"ScanEvalApp/internal/logging"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// SlicePdfForStudent slices a PDF file based on the pages specified in the students record in DB.
// It uses the pdftk tool to extract specific pages from the input PDF and saves the result to an output PDF.
func SlicePdfForStudent(db *gorm.DB, registrationNumber int) (string, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	student, err := latex.FindStudentByRegistrationNumber(db, registrationNumber)
	if err != nil {
		errorLogger.Error("Error finding student", "registration_number", registrationNumber, slog.String("error", err.Error()))
		return "", err
	}

	pagesStr := student.Pages

	if pagesStr == "" {
		err = fmt.Errorf("študent (číslo registrácie: %d) nemá žiadne stránky v databáze", registrationNumber)
		logger.Info("Študent nemá žiadne stránky v DB", "registration_number", registrationNumber)
		return "", err
	}

	pageParts := strings.Split(pagesStr, "-")
	var pages []int
	for _, part := range pageParts {
		if part == "" {
			continue // skipping empty parts
		}
		pageNum, err := strconv.Atoi(part)
		if err != nil {
			errorLogger.Error("Invalid page number in Pages", slog.String("value", part), slog.String("error", err.Error()))
			return "", err
		}
		pages = append(pages, pageNum)
	}

	logger.Info("Parsed pages", "registration_number", registrationNumber, "pages", pages)
	exam, err := repository.GetExam(db, student.ExamID)
	if err != nil {
		errorLogger.Error("Error retrieving exam", "exam_id", student.ExamID, slog.String("error", err.Error()))
		return "", err
	}

	safeTitle := common.SanitizeFilename(exam.Title)
	fileName := fmt.Sprintf("scan_%s_%d.pdf", safeTitle, exam.ID)
	inputPDF := filepath.Join(common.GLOBAL_TEMP_SCAN, fileName)

	if _, err := os.Stat(inputPDF); err != nil {
		if os.IsNotExist(err) {
			errorLogger.Error("PDF súbor pre test neexistuje", "file_path", inputPDF, slog.String("error", err.Error()))
			return "", fmt.Errorf("PDF súbor pre test neexistuje: %s", inputPDF)
		}
		errorLogger.Error("Chyba pri kontrole PDF súboru", "file_path", inputPDF, slog.String("error", err.Error()))
		return "", fmt.Errorf("chyba pri kontrole PDF súboru: %w", err)
	}

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
	outputPDF := filepath.Join(absDirPath, fmt.Sprintf("student_%d_vyplnene.pdf", registrationNumber))

	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p))
	}

	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	cmdArgs = append(cmdArgs, "output", outputPDF)

	cmd := exec.Command("pdftk", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorLogger.Error("Chyba pri spúšťaní pdftk", "error", err.Error(), "output", string(output))
		return "", err
	}

	logger.Info("PDF slicing pomocou pdftk hotový", "output_path", outputPDF)
	return outputPDF, nil
}

// ExportFailedPagesToPDF extracts a subset of pages (marked as failed) from the input PDF
// and saves them into a separate output PDF file.
func ExportFailedPagesToPDF(examTitle string, examID uint, pages []int, inputPDF string) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	if len(pages) == 0 {
		return nil
	}

	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p+1))
	}

	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	dirPath, err := config.LoadLastPath()
	if err != nil {
		errorLogger.Error("Chyba načítania configu", slog.String("error", err.Error()))
		return err
	}

	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		errorLogger.Error("Chyba pri konverzii cesty", slog.String("error", err.Error()))
		return err
	}

	outputPDF := filepath.Join(absDirPath, fmt.Sprintf("%s%d_failed_pages.pdf", examTitle, examID))
	cmdArgs = append(cmdArgs, "output", outputPDF)

	cmd := exec.Command("pdftk", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorLogger.Error("Chyba pri spájaní chybných stránok", "error", err.Error(), "output", string(output))
		return err
	}

	logger.Info("Chybné strany uložené do PDF", "output", outputPDF)
	return nil
}
