package seed

import "ScanEvalApp/internal/database/models"

func TestGenerator(questionsCount int, studentsCount int) *models.Test {
	test := &models.Test{
		Title:         "Matematický test",
		SchoolYear:    "2024/2025",
		QuestionCount: questionsCount,
		Questions:     generateAnswers(questionsCount),
		Students:      *StudentListGenerator(questionsCount, studentsCount),
		Date:			"1.1.2024",
	}
	return test
}
