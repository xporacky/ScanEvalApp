package main

import (
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"path/filepath"
	"text/template"
	"bytes"
)

func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
	// Create a temporary .tex file
	texFile, err := ioutil.TempFile("", "*.tex")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary LaTeX file: %v", err)
	}
	defer os.Remove(texFile.Name()) 

	// Write LaTeX content to the temporary file
	if _, err = texFile.Write(latexContent); err != nil {
		return nil, fmt.Errorf("error writing to .tex file: %v", err)
	}
	texFile.Close()

	// Create a temporary directory for the output
	outputDir, err := ioutil.TempDir("", "latex_output")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary output directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	// Run pdflatex to compile
	cmd := exec.Command("pdflatex", "-output-directory", outputDir, texFile.Name())
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("error compiling LaTeX to PDF: %v", err)
	}

	// Read the PDF file as bytes
	pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
	pdfBytes, err := ioutil.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("error reading PDF file: %v", err)
	}

	return pdfBytes, nil
}

// OpenFile načíta obsah súboru a vráti ho ako []byte
func OpenFile(filePath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("chyba pri otváraní súboru: %v", err)
	}
	return data, nil
}

// SaveFile uloží obsah []byte do súboru na špecifikovanej ceste
func SaveFile(filePath string, data []byte) error {
	err := ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("chyba pri ukladaní súboru: %v", err)
	}
	return nil
}


// LaTeX pdf generation

// štruktúra na prácu s nahrádzaním hodnôt v šablóne
type TemplateData struct {
	ID        string
	Meno      string
	Datum     string
	Miestnost string
	Cas       string
	Bloky     int
}

// Funkcia na nahradenie place holderov v LaTeX súbore
func ReplaceTemplatePlaceholders(templateContent []byte, data TemplateData) ([]byte, error) {
	tmpl, err := template.New("latex").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("chyba pri parsovaní šablóny: %v", err)
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, data)
	if err != nil {
		return nil, fmt.Errorf("chyba pri nahrádzaní hodnôt v šablóne: %v", err)
	}

	return output.Bytes(), nil
}



func main() {
	// Načítanie LaTeX súboru a jeho otvorenie
	latexFilePath := "./latexFiles/main.tex"
	latexContent, err := OpenFile(latexFilePath)
	if err != nil {
		fmt.Println("Error while opening LaTeX file:", err)
		return
	}

	// // Compile LaTeX to PDF
	// pdfBytes, err := CompileLatexToPDF(latexContent)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	
	// // Uloženie PDF
	// outputFilePath := "./tmp/output.pdf"
	// err = SaveFile(outputFilePath, pdfBytes)
	// if err != nil {
	// 	fmt.Println("Chyba pri ukladaní PDF súboru:", err)
	// 	return
	// }

	// fmt.Println("PDF úspešne vytvorený a uložený ako:", outputFilePath)

//	db, err := migrations.MigrateDB()
//	if err != nil {
//		panic("failed to connect to database")
//	}

//	seed.Seed(db)

//	fmt.Println("Database setup and seeding complete.")


	// LaTeX generation pdf test

	// Hodnoty na nahradenie kvôli testovaniu funkcionality
	data := TemplateData{
		ID:        "120345",
		Meno:      "Jožko Alexander Mrkvička",
		Datum:     "15. 11. 2024",
		Miestnost: "CD300",
		Cas:       "10:30",
		Bloky:     50,
	}
	

	// Nahradenie placeholderov
	updatedLatex, err := ReplaceTemplatePlaceholders(latexContent, data)
	if err != nil {
		fmt.Println("Error replacing placeholders:", err)
		return
	}	

	// Kompilácia upraveného LaTeX na PDF
	pdfBytes, err := CompileLatexToPDF(updatedLatex)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}	

	// Uloženie výsledného PDF
	outputFilePath := "./tmp/output.pdf"
	err = SaveFile(outputFilePath, pdfBytes)
	if err != nil {
		fmt.Println("Chyba pri ukladaní PDF súboru:", err)
		return
	}

	fmt.Println("PDF úspešne vytvorený a uložený ako:", outputFilePath)

}
