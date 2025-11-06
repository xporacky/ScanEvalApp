// In internal/database/migrations/migrations.go
package migrations

import (
"ScanEvalApp/internal/database/models"
"ScanEvalApp/internal/logging"
"log/slog"

"gorm.io/driver/sqlite"
"gorm.io/gorm"
)

func MigrateDB() (*gorm.DB, error) {
logger := logging.GetLogger()
errorLogger := logging.GetErrorLogger()

logger.Debug("Pripájam sa k databáze...")

db, err := gorm.Open(sqlite.Open("database/scan-eval-db.db"), &gorm.Config{})
if err != nil {
errorLogger.Error("Chyba pri pripájaní k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
return nil, err
}

logger.Debug("Spúšťam migrácie...")

// AutoMigrate will automatically add the OptionCount field to existing exams
err = db.AutoMigrate(&models.Exam{}, &models.Student{})
if err != nil {
errorLogger.Error("Chyba pri migrácii databázy", slog.Group("CRITICAL", slog.String("error", err.Error())))
return nil, err
}

// ⚠️ IMPORTANT: This ensures existing exams get option_count = 5
result := db.Model(&models.Exam{}).Where("option_count is NULL OR option_count = ?", 0).Update("option_count", 5)
if result.Error != nil {
errorLogger.Error("Chyba pri aktualizácii existujúcich testov", slog.String("error", result.Error.Error()))
} else if result.RowsAffected > 0 {
logger.Info("Aktualizované existujúce testy", slog.Int64("počet", result.RowsAffected))
}

logger.Info("Migrácia databázy úspešne dokončená")

return db, nil
}
