package latex

import (
	"ScanEvalApp/internal/database/models"
	"fmt"
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
