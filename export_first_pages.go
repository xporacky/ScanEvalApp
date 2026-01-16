package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
	"gocv.io/x/gocv"
)

func readQRFromMat(mat gocv.Mat) string {
	qrDetector := gocv.NewQRCodeDetector()
	points := gocv.NewMat()
	defer points.Close()
	qrCode := gocv.NewMat()
	defer qrCode.Close()
	text := qrDetector.DetectAndDecode(mat, &points, &qrCode)
	return text
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run export_first_pages.go <pdf_path>")
	}

	pdfPath := os.Args[1]
	outputDir := "student_pages"

	// Create output directory
	os.MkdirAll(outputDir, 0755)

	doc, err := fitz.New(pdfPath)
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	fmt.Printf("Exportujem prvé strany študentov z PDF (%d strán)\n\n", doc.NumPage())

	studentNum := 1

	// Process every 2 pages (each student has 2 pages)
	for pageNum := 0; pageNum < doc.NumPage(); pageNum += 2 {
		img, err := doc.Image(pageNum)
		if err != nil {
			log.Printf("Chyba pri načítaní strany %d: %v\n", pageNum+1, err)
			continue
		}

		// Convert to gocv.Mat to read QR
		mat, err := gocv.ImageToMatRGB(img)
		if err != nil {
			log.Printf("Chyba konverzie strany %d: %v\n", pageNum+1, err)
			continue
		}

		// Read QR code
		qrText := readQRFromMat(mat)
		mat.Close()

		// Save image with QR code in filename
		filename := fmt.Sprintf("student_%02d_page_%d_QR_%s.png", studentNum, pageNum+1, qrText)
		outputPath := filepath.Join(outputDir, filename)

		outFile, err := os.Create(outputPath)
		if err != nil {
			log.Printf("Chyba pri vytváraní súboru: %v\n", err)
			continue
		}

		err = png.Encode(outFile, img)
		outFile.Close()

		if err != nil {
			log.Printf("Chyba pri ukladaní obrázka: %v\n", err)
			continue
		}

		fmt.Printf("✓ Študent %2d (strana %2d) → QR=%3s → %s\n", studentNum, pageNum+1, qrText, filename)
		studentNum++
	}

	fmt.Printf("\n✅ Hotovo! Obrázky uložené v priečinku: %s\n", outputDir)
	fmt.Println("Teraz môžeš otvoriť tento priečinok a vidieť mená študentov na obrázkoch.")
}
