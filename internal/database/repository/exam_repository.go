package repository

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

func CreateExam(db *gorm.DB, test *models.Exam) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Vytváranie testu", slog.String("test", test.Title), slog.String("year", test.SchoolYear))
	result := db.Create(test)
	if result.Error != nil {
		errorLogger.Error("Chyba pri vytváraní testu", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return result.Error
}

func GetExam(db *gorm.DB, id uint) (*models.Exam, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Hľadanie testu", slog.Uint64("ID testu", uint64(id)))

	var test models.Exam
	result := db.First(&test, id)
	if result.Error != nil {
		errorLogger.Error("Test nebol nájdený", slog.Uint64("ID testu", uint64(id)), slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
		return nil, result.Error
	}
	logger.Debug("Test bol nájdený", slog.String("test", test.Title), slog.String("year", test.SchoolYear))
	return &test, nil
}

func GetAllExams(db *gorm.DB) ([]models.Exam, error) {
	errorLogger := logging.GetErrorLogger()

	var tests []models.Exam
	result := db.Preload("Students").Find(&tests) // Načítame aj priradených študentov
	if result.Error != nil {
		errorLogger.Error("Chyba pri načítavaní testov", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return tests, result.Error
}

func UpdateExam(db *gorm.DB, test *models.Exam) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Aktualizácia testu", slog.String("test", test.Title), slog.String("year", test.SchoolYear))
	result := db.Save(test)
	if result.Error != nil {
		errorLogger.Error("Chyba pri aktualizácii testu", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}
	return result.Error
}

func DeleteExam(db *gorm.DB, id uint) error {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	result := db.Delete(&models.Exam{}, id)
	if result.Error != nil {
		errorLogger.Error("Chyba pri mazaní testu", slog.Group("CRITICAL", slog.String("error", result.Error.Error())))
	}

	logger.Debug("Test vymazaný", slog.Uint64("test ID", uint64(id)))
	return result.Error
}

func ShowAnswers(test *models.Exam) {
	fmt.Println("Zobrazenie odpovedí na test: ")
	// Neskôr tu pridáme logiku na zobrazenie odpovedí.
}
