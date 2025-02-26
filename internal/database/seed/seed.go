package seed

import (
	"ScanEvalApp/internal/logging"
	"log/slog"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, questionCount int, studentsCount int) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	exam := ExamGenerator(questionCount, studentsCount)
	if err := db.Create(exam).Error; err != nil {
		errorLogger.Error("Could not seed exam", slog.Group("CRITICAL", slog.String("exam", exam.Title)))
	} else {
		logger.Info("Seeded exam", slog.String("exam", exam.Title))
	}
}

func GenerateAnswers(n int) string {
	answers := ""
	for i := 0; i < n; i++ {
		answers = answers + "0"
	}
	return answers
}
func RandomDate() time.Time {
	minUnix := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	maxUnix := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	randomUnix := rand.Int63n(maxUnix-minUnix) + minUnix
	return time.Unix(randomUnix, 0)
}
