package seed

import (
	"ScanEvalApp/internal/logging"
	"log/slog"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, questionCount int, studentsCount int) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	test := TestGenerator(questionCount, studentsCount)
	if err := db.Create(test).Error; err != nil {
		errorLogger.Error("Could not seed test", slog.Group("CRITICAL", slog.String("test", test.Title)))
	} else {
		logger.Info("Seeded test", slog.String("test", test.Title))
	}
}

func GenerateAnswers(n int) string {
	answers := ""
	for i := 0; i < n; i++ {
		answers = answers + "0"
	}
	return answers
}
