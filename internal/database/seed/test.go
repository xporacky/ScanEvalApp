package seed

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"log/slog"
)

func TestGenerator(questionsCount int, studentsCount int) *models.Test {
	logger := logging.GetLogger()

	logger.Debug("Generovanie testu", slog.Int("questions count", questionsCount), slog.Int("students count", studentsCount))

	test := &models.Test{
		Title:         "Matematický test",
		SchoolYear:    "2024/2025",
		Date:          RandomDate(),
		QuestionCount: questionsCount,
		Questions:     GenerateAnswers(questionsCount),
		Students:      *StudentListGenerator(questionsCount, studentsCount),
	}
	logger.Debug("Test úspešne vygenerovaný.")
	return test
}
