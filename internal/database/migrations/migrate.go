// In internal/database/migrations/migrations.go
package migrations

import (
	"os"
	"path/filepath"
	"strings"

	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const databaseFileName = "scan-eval-db.db"

func resolveDatabasePath() (string, error) {
	dbDir := filepath.Join("database")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", err
	}

	return filepath.Abs(filepath.Join(dbDir, databaseFileName))
}

func MigrateDB() (*gorm.DB, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Pripájam sa k databáze...")

	dbPath, err := resolveDatabasePath()
	if err != nil {
		errorLogger.Error("Chyba pri príprave cesty k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return nil, err
	}

	logger.Debug("Používam databázu", "path", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
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

	if err := ensureStudentRegistrationIndex(db); err != nil {
		errorLogger.Error("Chyba pri migrácii indexu študentov", slog.String("error", err.Error()))
		return nil, err
	}

	// ⚠️ IMPORTANT: This ensures existing exams get option_count = 5
	result := db.Model(&models.Exam{}).Where("option_count is NULL OR option_count = ?", 0).Update("option_count", 5)
	if result.Error != nil {
		errorLogger.Error("Chyba pri aktualizácii existujúcich testov", slog.String("error", result.Error.Error()))
	} else if result.RowsAffected > 0 {
		logger.Info("Aktualizované existujúce testy", slog.Int64("počet", result.RowsAffected))
	}

	showNameResult := db.Model(&models.Exam{}).Where("show_name IS NULL").Update("show_name", true)
	if showNameResult.Error != nil {
		errorLogger.Error("Chyba pri aktualizácii show_name pre existujúce testy", slog.String("error", showNameResult.Error.Error()))
	}

	logger.Info("Migrácia databázy úspešne dokončená")

	return db, nil
}

func ensureStudentRegistrationIndex(db *gorm.DB) error {
	logger := logging.GetLogger()

	var createTableSQL string
	if err := db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?", "students").Scan(&createTableSQL).Error; err != nil {
		return err
	}

	// Older schema used a global unique constraint on registration_number, which blocks
	// importing the same registration number into different exams.
	if strings.Contains(createTableSQL, "UNIQUE (`registration_number`)") {
		logger.Info("Migrujem students tabuľku na unikátnosť podľa exam_id + registration_number")

		tx := db.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		steps := []string{
			"PRAGMA foreign_keys = OFF",
			"ALTER TABLE students RENAME TO students_old",
			"CREATE TABLE students (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` text NOT NULL,`surname` text NOT NULL,`birth_date` datetime NOT NULL,`registration_number` integer NOT NULL,`room` text NOT NULL,`score` integer,`answers` text,`exam_id` integer NOT NULL,`pages` text,`exam_date` datetime DEFAULT NULL,`exam_time` text,`subgroup` text,CONSTRAINT `fk_exams_students` FOREIGN KEY (`exam_id`) REFERENCES `exams`(`id`))",
			"INSERT INTO students (`id`,`created_at`,`updated_at`,`deleted_at`,`name`,`surname`,`birth_date`,`registration_number`,`room`,`score`,`answers`,`exam_id`,`pages`,`exam_date`,`exam_time`,`subgroup`) SELECT `id`,`created_at`,`updated_at`,`deleted_at`,`name`,`surname`,`birth_date`,`registration_number`,`room`,`score`,`answers`,`exam_id`,`pages`,`exam_date`,`exam_time`,`subgroup` FROM students_old",
			"DROP TABLE students_old",
			"CREATE INDEX IF NOT EXISTS `idx_students_deleted_at` ON `students`(`deleted_at`)",
			"CREATE UNIQUE INDEX IF NOT EXISTS `idx_students_exam_registration` ON `students`(`exam_id`,`registration_number`)",
			"PRAGMA foreign_keys = ON",
		}

		for _, stmt := range steps {
			if err := tx.Exec(stmt).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		return tx.Commit().Error
	}

	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS `idx_students_exam_registration` ON `students`(`exam_id`,`registration_number`)").Error
}
