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

// removeDiacritics odstráni diakritiku a nahradí medzery znakom _
func removeDiacritics(input string) string {
	// Normalizuje text do NFD formy (rozklad na základný znak + diakritické znamienko)
	t := norm.NFD.String(input)
	// Odstráni všetky neascii znaky (diakritiku)
	t = strings.Map(func(r rune) rune {
		if unicode.IsMark(r) {
			return -1
		}
		return r
	}, t)
	// Nahradí medzery podtržníkom
	t = strings.ReplaceAll(t, " ", "_")
	return t
}

// FindStudentByRegistrationNumber nájde študenta v DB podľa RegistrationNumber
func FindStudentByRegistrationNumber(db *gorm.DB, registrationNumber int) (*models.Student, error) {
	var student models.Student
	if err := db.Where("registration_number = ?", registrationNumber).First(&student).Error; err != nil {
		return nil, fmt.Errorf("student not found with RegistrationNumber %d: %w", registrationNumber, err)
	}
	return &student, nil
}

func SlicePdfForStudent(db *gorm.DB, registrationNumber int) error {
	// Inicializácia loggera
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Najprv nájdi študenta podľa RegistrationNumber
	student, err := FindStudentByRegistrationNumber(db, registrationNumber)
	if err != nil {
		errorLogger.Error("Error finding student", "registration_number", registrationNumber, slog.String("error", err.Error()))
		return err
	}

	// rozparsuj stlpec pages cez delimiter -
	pagesStr := student.Pages

	if pagesStr == "" {
		logger.Info("Študent nemá žiadne stránky v DB", "registration_number", registrationNumber)
		return nil // TODO osetrit lebo v GUI sa zasa ked sa nevrati error vypise ze to bolo uspesne slicnutie
	}

	pageParts := strings.Split(pagesStr, "-")

	var pages []int
	for _, part := range pageParts {
		if part == "" {
			continue // pre istotu preskočíme prázdne
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
	outputPDF := filepath.Join(OutputPDFPath, fmt.Sprintf("student_%d_vyplnene.pdf", registrationNumber))

	// Vytvor string zo strán: napr. "1 3 5"
	var pageArgs []string
	for _, p := range pages {
		pageArgs = append(pageArgs, strconv.Itoa(p))
	}

	// Vytvor exec.Command argumenty ako samostatné stringy
	cmdArgs := append([]string{inputPDF, "cat"}, pageArgs...)
	cmdArgs = append(cmdArgs, "output", outputPDF)

	// Spusti pdftk s rozdelenými argumentmi
	cmd := exec.Command("pdftk", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorLogger.Error("Chyba pri spúšťaní pdftk", "error", err.Error(), "output", string(output))
		return err
	}

	logger.Info("PDF slicing pomocou pdftk hotový", "output_path", outputPDF)
	return nil
}
