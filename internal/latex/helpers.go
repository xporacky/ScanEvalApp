package latex

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"

	"log/slog"

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
