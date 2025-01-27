package main

import (
	"ScanEvalApp/internal/database/migrations"
	"ScanEvalApp/internal/database/seed"
	"ScanEvalApp/internal/latex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Kontrola ci je databaza prazdna
func CheckIfDatabaseIsEmpty(db *gorm.DB) (bool, error) {
	var count int64

	err := db.Table("students").Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// pomocna funkcia, ktora robi seedovanie, pokial mame prazdnu databazu (kvoli testovaniu generovania pdf pre studentov)
func testDatabase(questionsCount int) {
	db, err := migrations.MigrateDB()
	if err != nil {
		panic("failed to connect to database")
	}

	isEmpty, err := CheckIfDatabaseIsEmpty(db)
	if err != nil {
		fmt.Println("Error while checking database:", err)
		return
	}

	if isEmpty {
		seed.Seed(db, questionsCount)
		fmt.Println("Database was empty. Seeding complete.")
	} else {
		fmt.Println("Database is not empty. Skipping seeding.")
	}
}

func main() {
	var questionsCount int = 40 // pocet otazok ktore pridelujeme do testu, na testovanie

	// inicializacia a migracia db
	db, err := migrations.MigrateDB()
	if err != nil {
		panic("failed to connect to database")
	}

	// kontrola ci je prazdna databaza, kvoli seedovaniu
	testDatabase(questionsCount)

	// Cesta k LaTeX sablone a cesta kam sa ma ulozit finalne pdf (pdf so studentami)
	templatePath := "./assets/latex/main.tex"
	outputPDFPath := "./assets/tmp/final.pdf"

	// Generovanie PDF pre vsetkych studentov
	if err := latex.ParallelGeneratePDFs(db, templatePath, outputPDFPath); err != nil {
		fmt.Println("Chyba pri generovaní PDF:", err)
		return
	}

	start := time.Now()
	//test := repository.GetTest(db, 2)
	//scanprocessing.ProcessPDF("assets/tmp/scan-pdfs", "assets/tmp/scan-images", test, db)
	elapsed := time.Since(start)
	fmt.Printf("Function took %s\n", elapsed)
}
