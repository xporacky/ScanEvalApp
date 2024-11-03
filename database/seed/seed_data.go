package seed

import (
	"time"
	"gorm.io/gorm"
	"ScanEvalApp/database/models"
	"ScanEvalApp/database/repository"
)

func SeedTestData(db *gorm.DB) {
	test := models.Test{
		Title:         "Matematický test",
		SchoolYear:    "2024/2025",
		QuestionCount: 3,
		Questions:     "abc",
		Students: []models.Student{
			{
				Name:               "Ján",
				Surname:            "Novák",
				BirthDate:          time.Date(2001, 5, 15, 0, 0, 0, 0, time.UTC),
				RegistrationNumber: "20210001",
				Room:               "A101",
				Score:              85,
				Answers:            "abc",
			},
		},
	}

	err := repository.CreateTest(db, &test)
	if err != nil {
		panic("Failed to seed test data")
	}
}
