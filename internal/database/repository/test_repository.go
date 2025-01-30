package repository

import (
	"ScanEvalApp/internal/database/models"
	"fmt"
	"gorm.io/gorm"
)

func CreateTest(db *gorm.DB, test *models.Test) error {
	result := db.Create(test)
	return result.Error
}

func GetTest(db *gorm.DB, id uint) (*models.Test, error) {
	var test models.Test
	result := db.First(&test, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &test, nil
}

func GetAllTests(db *gorm.DB) ([]models.Test, error) {
	var tests []models.Test
	result := db.Find(&tests)
	return tests, result.Error
}

func UpdateTest(db *gorm.DB, test *models.Test) error {
	result := db.Save(test)
	return result.Error
}

func DeleteTest(db *gorm.DB, id uint) error {
	result := db.Delete(&models.Test{}, id)
	return result.Error
}

func ShowAnswers(test *models.Test) {
	fmt.Println("Zobrazenie odpovedí na test: ")
	// Neskôr tu pridáme logiku na zobrazenie odpovedí.
}