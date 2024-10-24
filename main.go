package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CompileLatexToPDF(latexFilePath string) error {
	if _, err := os.Stat(latexFilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", latexFilePath)
	}

	cmd := exec.Command("pdflatex", latexFilePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to compile LaTeX file: %v", err)
	}

	return nil
}
func CreateTest(db *gorm.DB) {
	// Vytvorenie nového testu s otázkami a študentmi
	test := Test{
		Title:         "Matematický test",
		SchoolYear:    "2024/2025",
		QuestionCount: 3,
		Questions: map[int]rune{
			1: 'A',
			2: 'B',
			3: 'C',
		},
		Students: []Student{
			{
				Name:              "Ján",
				Surname:           "Novák",
				BirthDate:         time.Date(2001, 5, 15, 0, 0, 0, 0, time.UTC),
				RegistrationNumber: "20210001",
				Room:              "A101",
				Score:             85,
				Answers: map[int]rune{
					1: 'A',
					2: 'B',
					3: 'C',
				},
			},
		},
	}

	// Uloženie testu a študentov do databázy
	result := db.Create(&test)
	if result.Error != nil {
		panic("Failed to create test")
	}
}

func main() {
/*	latexFilePath := "./latexFiles/main.tex"
	err := CompileLatexToPDF(latexFilePath)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("PDF compiled successfully.")
	}
*/
	db, err := gorm.Open(sqlite.Open("scan-eval-db.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	err = db.AutoMigrate(&Test{}, &Student{})
	if err != nil {
		panic("failed to migrate database")
	}

	CreateTest(db)
}
