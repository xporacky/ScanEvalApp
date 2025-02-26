package seed

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/logging"
	"log/slog"
)

func ExamGenerator(questionsCount int, studentsCount int) *models.Exam {
	logger := logging.GetLogger()

	logger.Debug("Generovanie testu", slog.Int("questions count", questionsCount), slog.Int("students count", studentsCount))

	exam := &models.Exam{
		Title:         "Matematický test",
		SchoolYear:    "2024/2025",
		Date:          RandomDate(),
		QuestionCount: questionsCount,
		Questions:     GenerateAnswers(questionsCount),
		Students:      *StudentListGenerator(questionsCount, studentsCount),
	}
	logger.Debug("Test úspešne vygenerovaný.")
	return exam
}
