package migrations

import (
	"ScanEvalApp/internal/database/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"ScanEvalApp/internal/logging"
	"log/slog"
)

func MigrateDB() (*gorm.DB, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Pripájam sa k databáze...")

	db, err := gorm.Open(sqlite.Open("internal/database/scan-eval-db.db"), &gorm.Config{})
	if err != nil {
		errorLogger.Error("Chyba pri pripájaní k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	logger.Debug("Spúšťam migrácie...")

	err = db.AutoMigrate(&models.Test{}, &models.Student{})
	if err != nil {
		errorLogger.Error("Chyba pri migrácii databázy", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	return db, nil
}
