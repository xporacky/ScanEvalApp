package latex

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"

	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strconv"

	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

// removeDiacritics removes diacritics from the input string and replaces spaces with underscores.
// It normalizes the string to NFD form, removes diacritical marks, and replaces spaces with underscores.
func removeDiacritics(input string) string {
	t := norm.NFD.String(input)
	t = strings.Map(func(r rune) rune {
		if unicode.IsMark(r) {
			return -1
		}
		return r
	}, t)
	// Replace spaces with underscores
	t = strings.ReplaceAll(t, " ", "_")
	return t
}

// FindStudentByRegistrationNumber finds a student in the database by their registration number.
// It returns a pointer to the student if found, or an error if not found.
func FindStudentByRegistrationNumber(db *gorm.DB, registrationNumber int) (*models.Student, error) {
	errorLogger := logging.GetErrorLogger()
	var student models.Student
	// Query the database for the student with the specified registration number
	if err := db.Where("registration_number = ?", registrationNumber).First(&student).Error; err != nil {
		// Log an error if the student is not found
		errorLogger.Error("Student not found with ", slog.Uint64("registration_number", uint64(registrationNumber)), slog.String("error", err.Error()))
		return nil, err
	}
	return &student, nil
}

// SlicePdfForStudent slices a PDF file based on the pages specified in the students record in DB.
// It uses the pdftk tool to extract specific pages from the input PDF and saves the result to an output PDF.
func SlicePdfForStudent(db *gorm.DB, registrationNumber int) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	student, err := FindStudentByRegistrationNumber(db, registrationNumber)
	if err != nil {
		errorLogger.Error("Error finding student", "registration_number", registrationNumber, slog.String("error", err.Error()))
		return err
	}

	// Parse the pages string from the student record and split it into individual pages
	pagesStr := student.Pages

	if pagesStr == "" {
		logger.Info("Študent nemá žiadne stránky v DB", "registration_number", registrationNumber)
		return nil // TODO osetrit lebo v GUI sa zasa ked sa nevrati error vypise ze to bolo uspesne slicnutie
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
			return err
		}
		pages = append(pages, pageNum)
	}

	logger.Info("Parsed pages", "registration_number", registrationNumber, "pages", pages)

	// TODO - zmenit staticku cestu, treba vybrat dynamicky z priecinka
	inputPDF := "/home/timo/ScanEvalApp/assets/tmp/scan-pdfs/sken_zasadacka_190_400dpi.pdf"
	outputPDF := filepath.Join(OUTPUT_PDF_PATH, fmt.Sprintf("student_%d_vyplnene.pdf", registrationNumber))

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
		return err
	}

	logger.Info("PDF slicing pomocou pdftk hotový", "output_path", outputPDF)
	return nil
}
