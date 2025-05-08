package pdf

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/latex"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/logging"
	"fmt"
	"log/slog"
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

	// Parse the pages string from the student record and split it into individual pages
	pagesStr := student.Pages

	if pagesStr == "" {
		logger.Info("Študent nemá žiadne stránky v DB", "registration_number", registrationNumber)
		return "", nil // TODO osetrit lebo v GUI sa zasa ked sa nevrati error vypise ze to bolo uspesne slicnutie
	}

	// Split the pages string by the delimiter "-" and convert to integer values
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
	exam, _:= repository.GetExam(db, student.ExamID)
	fileName := fmt.Sprintf("scan_%s_%d.pdf", exam.Title, exam.ID)
	// TODO - zmenit staticku cestu, treba vybrat dynamicky z priecinka
	//inputPDF := "/home/timo/ScanEvalApp/assets/tmp/scan-pdfs/sken_zasadacka_190_400dpi.pdf"

	inputPDF := filepath.Join(common.GLOBAL_TEMP_SCAN, fileName)

	outputPDF := filepath.Join(common.GLOBAL_EXPORT_DIR, fmt.Sprintf("student_%d_vyplnene.pdf", registrationNumber))

	// Convert the list of pages into arguments for pdftk
	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p))
	}

	// Vytvor exec.Command argumenty ako samostatné stringy
	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	cmdArgs = append(cmdArgs, "output", outputPDF)

	// Prepare the command-line arguments for pdftk
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
	outputPDF := filepath.Join(common.GLOBAL_EXPORT_DIR, fmt.Sprintf("%s%d_failed_pages.pdf", examTitle, examID))
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
