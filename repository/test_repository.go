package repository

import (
	"gorm.io/gorm"
	"ScanEvalApp/models"
)

func CreateTest(db *gorm.DB, test *models.Test) error {
	result := db.Create(test)
	return result.Error
}
