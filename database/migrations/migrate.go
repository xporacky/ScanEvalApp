package migrations

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"ScanEvalApp/database/models"
)

func MigrateDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("scan-eval-db.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Test{}, &models.Student{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
