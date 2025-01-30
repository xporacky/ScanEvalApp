package seed

import (
	"fmt"
	"math/rand"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, questionCount int, studentsCount int) {
	test := TestGenerator(questionCount, studentsCount)
	if err := db.Create(test).Error; err != nil {
		fmt.Printf("Could not seed test: %s\n", test.Title)
	} else {
		fmt.Printf("Seeded test: %s\n", test.Title)
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
