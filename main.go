package main

import (
	"ScanEvalApp/internal/database/migrations"
	"ScanEvalApp/internal/database/seed"

	//"ScanEvalApp/internal/files"
	//"ScanEvalApp/internal/latex"
	//"ScanEvalApp/internal/scanprocessing"
	window "ScanEvalApp/internal/gui"
	"ScanEvalApp/internal/logging"

	"gioui.org/app"

	//"fmt"
	//"time"

	"log/slog"

	"gorm.io/gorm"
)

// Kontrola ci je databaza prazdna
func CheckIfDatabaseIsEmpty(db *gorm.DB) (bool, error) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	var count int64

	err := db.Table("students").Count(&count).Error
	if err != nil {
		errorLogger.Error("Chyba pri kontrole databázy", slog.String("error", err.Error()))
		return false, err
	}
	if count == 0 {
		logger.Warn("Databáza je prázdna.")
	} else {
		logger.Info("Databáza obsahuje záznamy.", slog.Int64("count", count))
	}
	return count == 0, nil
}

// pomocna funkcia, ktora robi seedovanie, pokial mame prazdnu databazu (kvoli testovaniu generovania pdf pre studentov)
func testDatabase(questionsCount int, studentsCount int) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Debug("Inicializujem databázu na testovanie.")

	db, err := migrations.MigrateDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error()))) // Použi Error namiesto Critical
		panic("failed to connect to database")
	}

	isEmpty, err := CheckIfDatabaseIsEmpty(db)
	if err != nil {
		errorLogger.Error("Chyba pri kontrole prázdnosti databázy", slog.String("error", err.Error()))
		return
	}

	if isEmpty {
		logger.Info("Databáza je prázdna, spúšťam seedovanie.")
		seed.Seed(db, questionsCount, studentsCount)
		logger.Info("Seedovanie dokončené.")
	} else {
		logger.Info("Databáza nie je prázdna, seedovanie preskočené.")
	}
}

func main() {
	logging.InitLogger()
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	logger.Info("---------------------------------------------------")
	errorLogger.Error("---------------------------------------------------")

	logger.Info("Aplikácia spustená")

	var questionsCount int = 40 // pocet otazok ktore pridelujeme do testu, na testovanie
	var studentsCount int = 50

	// inicializacia a migracia db
	logger.Info("Spúšťam migráciu databázy.")
	db, err := migrations.MigrateDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error()))) // Použi Error namiesto Critical
		panic("failed to connect to database")
	}
	logger.Info("Migrácia databázy dokončená.")

	// kontrola ci je prazdna databaza, kvoli seedovaniu
	testDatabase(questionsCount, studentsCount)
	/*
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
	*/

	logger.Info("Spúšťam GUI.")
	go window.RunWindow(db) // Zavolanie funkcie na vytvorenie a správu okna
	app.Main()
}
