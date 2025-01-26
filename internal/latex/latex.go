package latex

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/gorm"

	"ScanEvalApp/internal/database/models"
)

// funkcia na kompilaciu LaTeX sablony do PDF
func CompileLatexToPDF(latexContent []byte) ([]byte, error) {
	texFile, err := os.CreateTemp("", "*.tex")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary LaTeX file: %v", err)
	}
	defer os.Remove(texFile.Name())

	if _, err = texFile.Write(latexContent); err != nil {
		return nil, fmt.Errorf("error writing to .tex file: %v", err)
	}
	texFile.Close()

	outputDir, err := os.MkdirTemp("", "latex_output")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary output directory: %v", err)
	}
	defer os.RemoveAll(outputDir)

	cmd := exec.Command("pdflatex", "-output-directory", outputDir, texFile.Name())
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("error compiling LaTeX to PDF: %v", err)
	}

	pdfPath := filepath.Join(outputDir, filepath.Base(texFile.Name())[:len(filepath.Base(texFile.Name()))-4]+".pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("error reading PDF file: %v", err)
	}

	return pdfBytes, nil
}

// funkcia na nahradenie placeholderov v LaTeX sablone
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

// mergovanie 2 pdf pomocou pdfunite kniznice
func MergePDFs(pdf1Path, pdf2Path, outputPath string) error {
	cmd := exec.Command("pdfunite", pdf1Path, pdf2Path, outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error merging PDFs: %v", err)
	}
	return nil
}

// funkcia na paralelne generovanie pdf
func ParallelGeneratePDFs(db *gorm.DB, templatePath, outputPDFPath string) error {
	// nacitanie vsetkych studentov z databazy
	var students []models.Student
	if err := db.Find(&students).Error; err != nil {
		return fmt.Errorf("error fetching students: %v", err)
	}

	mainPDFPath := filepath.Join(os.TempDir(), "main.pdf")
	initialPDF := []byte("%PDF-1.4\n%âãÏÓ\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Count 1 /Kids [3 0 R] >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 44 >>\nstream\n0 0 0 1 0 0 0 re\nf 0 0 0 1 0 0 0 Tm /F1 12 Tf 72 720 Td (Test) Tj\nendstream\nendobj\ntrailer\n<< /Root 1 0 R /Size 5 >>\n%%EOF")
	if err := os.WriteFile(mainPDFPath, initialPDF, 0644); err != nil {
		return fmt.Errorf("error creating initial PDF: %v", err)
	}

	// meranie celkoveho casu generovania a mergovania pdf
	startTime := time.Now()

	// synchronizacia prace s goroutines pomocou WaitGroup
	var wg sync.WaitGroup
	var pdfMergeMutex sync.Mutex
	var processedCount int64
	totalStudents := len(students)

	// paralelne generovanie pdf pre studentov
	for _, student := range students {
		wg.Add(1)
		go func(student models.Student) {
			defer wg.Done()

			// meranie casu spracovania generovanie pdf pre studenta
			studentStartTime := time.Now()

			// Nacitanie LaTeX sablony
			latexTemplate, err := os.ReadFile(templatePath)
			if err != nil {
				fmt.Printf("Error reading LaTeX template for student %d: %v\n", student.ID, err)
				return
			}

			// nacitanie testu pre studenta z databazy
			var test models.Test
			if err := db.First(&test, student.TestID).Error; err != nil {
				fmt.Printf("Error fetching test for student %d: %v\n", student.ID, err)
				return
			}

			// vytvorenie dat, ktore budu nacitane namiesto placeholderov v LaTeX sablone
			data := TemplateData{
				ID:        fmt.Sprintf("%d", student.ID),
				Meno:      fmt.Sprintf("%s %s", student.Name, student.Surname),
				Datum:     "25. 1. 2025", // TODO -> pri vytvarani testu tiez bude musiet admin zadat datum toho testu
				Miestnost: student.Room,
				Cas:       "10:30", // TODO - beh treba dopocitat z noveho stlpca (Beh testu (budu asi 2 behy, a z kazdeho je vzdy jasny cas, napr 1.beh = 10:00))
				Bloky:     test.QuestionCount,
				QrCode:    fmt.Sprintf("%d", student.ID),
			}

			// nahradenie placeholderov v LaTeX sablone udajmi studenta
			updatedLatex, err := ReplaceTemplatePlaceholders(latexTemplate, data)
			if err != nil {
				fmt.Printf("Error replacing placeholders for student %d: %v\n", student.ID, err)
				return
			}

			// kompilacia latex sablony do pdf studenta
			studentPDF, err := CompileLatexToPDF(updatedLatex)
			if err != nil {
				fmt.Printf("Error generating PDF for student %d: %v\n", student.ID, err)
				return
			}

			// ulozenie generovaneho/kompilovaneho pdf studenta do docasneho suboru
			studentPDFPath := filepath.Join(os.TempDir(), fmt.Sprintf("student_%d.pdf", student.ID))
			if err := os.WriteFile(studentPDFPath, studentPDF, 0644); err != nil {
				fmt.Printf("Error saving PDF for student %d: %v\n", student.ID, err)
				return
			}

			// zamok na zlucovanie pdf, aby sa predislo chybam pri zlucovani
			pdfMergeMutex.Lock()
			defer pdfMergeMutex.Unlock()

			// zlucenie generovaneho pdf so zakladnym pdf
			mergedPDFPath := filepath.Join(os.TempDir(), "merged.pdf")
			if err := MergePDFs(mainPDFPath, studentPDFPath, mergedPDFPath); err != nil {
				fmt.Printf("Error merging PDF for student %d: %v\n", student.ID, err)
				return
			}

			// nahradenie hlavneho pdf novym zlucenym pdf
			if err := os.Rename(mergedPDFPath, mainPDFPath); err != nil {
				fmt.Printf("Error updating main PDF for student %d: %v\n", student.ID, err)
				return
			}

			// Odstranenie docasneho pdf studenta
			if err := os.Remove(studentPDFPath); err != nil {
				fmt.Printf("Error removing temporary PDF for student %d: %v\n", student.ID, err)
				return
			}

			// Zvýšenie počtu spracovaných PDF a výpis stavu
			processedCount++
			studentDuration := time.Since(studentStartTime)
			fmt.Printf("(%d/%d) Generovanie PDF (Test: %s) s id študenta: %d, dokončené za: %v\n", processedCount, totalStudents, test.Title, student.ID, studentDuration)
		}(student)
	}

	// cakanie na vsetky goroutines
	wg.Wait()

	// odstranenie prvej stranky pdf (bez prazdnej vygenerovanej strany to robilo problemy pri mergovani)
	finalPDFPath := filepath.Join(os.TempDir(), "final_output.pdf")
	cmdRemoveFirstPage := exec.Command("pdftk", mainPDFPath, "cat", "2-end", "output", finalPDFPath)
	if err := cmdRemoveFirstPage.Run(); err != nil {
		return fmt.Errorf("error removing first page: %v", err)
	}

	// presun finalneho pdf na zadanu cestu
	if err := os.Rename(finalPDFPath, outputPDFPath); err != nil {
		return fmt.Errorf("error moving final PDF: %v", err)
	}

	// zmeranie celkoveho casu behu programu
	duration := time.Since(startTime)
	fmt.Printf("Celkový čas generovania PDF: %v\n", duration)

	fmt.Printf("Výsledné PDF úspešne uložené do: %s\n", outputPDFPath)
	return nil
}
