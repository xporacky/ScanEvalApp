package repository

import (
	"gorm.io/gorm"
	"ScanEvalApp/database/models"
)

func CreateTest(db *gorm.DB, test *models.Test) error {
	result := db.Create(test)
	return result.Error
}
