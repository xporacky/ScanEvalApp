package seed

import (
	"math/rand"

	"gorm.io/gorm"
	"ScanEvalApp/internal/logging"
	"log/slog"
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

func generateAnswers(n int) string {
	possibleAnswers := []string{"A", "B", "C", "D", "E"}
	answers := ""
	for i := 0; i < n; i++ {
		answers = answers + possibleAnswers[rand.Intn(len(possibleAnswers))]
	}
	return answers
}
