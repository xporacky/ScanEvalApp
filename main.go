package main

import (
	"ScanEvalApp/internal/database/migrations"
	"ScanEvalApp/internal/database/seed"

	//"ScanEvalApp/internal/files"
	//"ScanEvalApp/internal/latex"
	//"ScanEvalApp/internal/scanprocessing"
	window "ScanEvalApp/internal/gui"

	"gioui.org/app"

	"fmt"
	//"time"
)

func testDatabase() {
	db, err := migrations.MigrateDB()
	if err != nil {
		panic("failed to connect to database")
	}

	seed.Seed(db)

	fmt.Println("Database setup and seeding complete.")
}
func main() {
	go window.RunWindow() // Zavolanie funkcie na vytvorenie a správu okna
	app.Main()

	/*
		// Načítanie LaTeX súboru a jeho otvorenie

		latexFilePath := "./assets/latex/main.tex"
		latexContent, err := files.OpenFile(latexFilePath)
		if err != nil {
			fmt.Println("Error while opening LaTeX file:", err)
			return
		}

		//	db, err := migrations.MigrateDB()
		//	if err != nil {
		//		panic("failed to connect to database")
		//	}

		//	seed.Seed(db)

		//	fmt.Println("Database setup and seeding complete.")

		// LaTeX generation pdf test
		// Hodnoty na nahradenie kvôli testovaniu funkcionality
		data := latex.TemplateData{
			ID:        "120345",
			Meno:      "Jožko Alexander Mrkvička",
			Datum:     "15. 11. 2024",
			Miestnost: "CD300",
			Cas:       "10:30",
			Bloky:     50,
			QrCode:    "www.google.com",
		}

		// Nahradenie placeholderov
		updatedLatex, err := latex.ReplaceTemplatePlaceholders(latexContent, data)
		if err != nil {
			fmt.Println("Error replacing placeholders:", err)
			return
		}

		// Kompilácia upraveného LaTeX na PDF
		pdfBytes, err := latex.CompileLatexToPDF(updatedLatex)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Uloženie výsledného PDF
		outputFilePath := "./assets/tmp/output.pdf"
		err = files.SaveFile(outputFilePath, pdfBytes)
		if err != nil {
			fmt.Println("Chyba pri ukladaní PDF súboru:", err)
			return
		}

		fmt.Println("PDF úspešne vytvorený a uložený ako:", outputFilePath)
		start := time.Now()
		//test := repository.GetTest(db, 2)
		//scanprocessing.ProcessPDF("assets/tmp/scan-pdfs", "assets/tmp/scan-images", test, db)
		elapsed := time.Since(start)
		fmt.Printf("Function took %s\n", elapsed)
	*/
}
