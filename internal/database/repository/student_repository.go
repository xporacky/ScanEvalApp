package repository

import (
	"ScanEvalApp/internal/database/models"
	"fmt"

	"gorm.io/gorm"
	//"fmt"
	"ScanEvalApp/internal/logging"
	"log/slog"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func CreateStudent(db *gorm.DB, student *models.Student) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Vytváranie študenta", slog.String("name", student.Name), slog.String("surname", student.Surname), "registration number", student.RegistrationNumber)
	result := db.Create(student)
	if result.Error != nil {
		errorLogger.Error("Chyba pri vytváraní študenta", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return result.Error
}

func GetStudent(db *gorm.DB, registrationNumber uint, testID uint) (*models.Student, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Hľadanie študenta", slog.Uint64("registration number", uint64(registrationNumber)), slog.Uint64("test ID", uint64(testID)))

	var student models.Student
	result := db.Where("ID = ? AND test_id = ?", registrationNumber, testID).First(&student)
	if result.Error != nil {
		errorLogger.Error("Študent nebol nájdený", "student registration number", student.RegistrationNumber, slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
		return nil, result.Error
	}
	logger.Debug("Študent nájdený", slog.String("name", student.Name), slog.String("surname", student.Surname))
	return &student, nil
}

func GetAllStudents(db *gorm.DB) ([]models.Student, error) {
	errorLogger := logging.GetErrorLogger()

	var students []models.Student
	result := db.Find(&students)
	if result.Error != nil {
		errorLogger.Error("Chyba pri načítavaní študentov", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return students, result.Error
}

func UpdateStudent(db *gorm.DB, student *models.Student) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Aktualizácia študenta", "registration number", student.RegistrationNumber, slog.String("name", student.Name), slog.String("surname", student.Surname))
	result := db.Save(student)
	if result.Error != nil {
		errorLogger.Error("Chyba pri aktualizácii študenta", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return result.Error
}

func DeleteStudent(db *gorm.DB, student *models.Student) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	logger.Debug("Mazanie študenta", "registration number", student.RegistrationNumber, slog.String("name", student.Name), slog.String("surname", student.Surname))
	result := db.Delete(student)
	if result.Error != nil {
		errorLogger.Error("Chyba pri mazaní študenta", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}

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
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Vyhľadávanie študentov podľa dotazu", slog.String("query", query))

	var students []models.Student
	query = removeDiacritics(query) // Odstráni diakritiku

	rows, err := db.Raw("SELECT * FROM students").Rows()
	if err != nil {
		errorLogger.Error("Chyba pri vyhľadávaní študentov", slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		db.ScanRows(rows, &student)
		// Porovnanie bez diakritiky
		if strings.Contains(removeDiacritics(student.Name), query) ||
			strings.Contains(removeDiacritics(student.Surname), query) ||
			strings.Contains(fmt.Sprintf("%d", student.RegistrationNumber), query) {
			students = append(students, student)
		}
	}
	logger.Info("Počet nájdených študentov", slog.Int("count", len(students)))

	return students, nil
}
