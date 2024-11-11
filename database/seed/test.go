package seed

import (
	"ScanEvalApp/database/models"
)

func TestGenerator(questionsCount int) *models.Test {
	test := &models.Test{
		Title:         "Matematický test",
		SchoolYear:    "2024/2025",
		QuestionCount: 3,
		Questions:     generateAnswers(questionsCount),
		Students:      *StudentListGenerator(questionsCount),
	}
	return test
}
