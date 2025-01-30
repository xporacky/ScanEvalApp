package repository

import (
	"ScanEvalApp/internal/database/models"

	"gorm.io/gorm"
	//"fmt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

func CreateStudent(db *gorm.DB, student *models.Student) error {
	result := db.Create(student)
	return result.Error
}

func GetStudent(db *gorm.DB, registrationNumber uint, testID uint) (*models.Student, error) {
	var student models.Student
	result := db.Where("registration_number = ? AND test_id = ?", registrationNumber, testID).First(&student)
	if result.Error != nil {
		return nil, result.Error
	}
	return &student, nil
}

func GetAllStudents(db *gorm.DB) ([]models.Student, error) {
	var students []models.Student
	result := db.Find(&students)
	return students, result.Error
}

func UpdateStudent(db *gorm.DB, student *models.Student) error {
	result := db.Save(student)
	return result.Error
}

func DeleteStudent(db *gorm.DB, id uint) error {
	result := db.Delete(&models.Student{}, id)
	return result.Error
}

// Funkcia na odstránenie diakritiky
func removeDiacritics(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Odstráni diakritické značky
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return strings.ToLower(result) // Konvertuje na malé písmená pre case-insensitive porovnávanie
}

func GetStudentsQuery(db *gorm.DB, query string) ([]models.Student, error) {
	var students []models.Student
	query = removeDiacritics(query) // Odstráni diakritiku

	rows, err := db.Raw("SELECT * FROM students").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		db.ScanRows(rows, &student)
		// Porovnanie bez diakritiky
		if strings.Contains(removeDiacritics(student.Name), query) ||
			strings.Contains(removeDiacritics(student.Surname), query) ||
			strings.Contains(student.RegistrationNumber, query) {
			students = append(students, student)
		}
	}
	return students, nil
}
