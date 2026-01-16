package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gen2brain/go-fitz"
	"gocv.io/x/gocv"
)

func extractNameOCR(mat gocv.Mat) string {
	// Crop header area where name is (top portion of the page)
	headerHeight := 400
	if mat.Rows() < 400 {
		headerHeight = mat.Rows() / 3
	}

	rect := image.Rectangle{
		Min: image.Point{100, 100},
		Max: image.Point{mat.Cols() - 100, headerHeight},
	}

	headerMat := mat.Region(rect)
	defer headerMat.Close()

	// Save temporarily
	tmpFile := "/tmp/header_ocr.png"
	gocv.IMWrite(tmpFile, headerMat)
	defer os.Remove(tmpFile)

	// Run Tesseract with Slovak language
	cmd := exec.Command("tesseract", tmpFile, "stdout", "-l", "slk+eng", "--psm", "6")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	text := string(output)
	lines := strings.Split(text, "\n")

	// Look for name - typically appears after "Meno:" label
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Check if line contains "Meno:"
		if strings.Contains(strings.ToLower(line), "meno") {
			// Name might be on same line after colon
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					name := strings.TrimSpace(parts[1])
					if len(name) > 3 {
						return cleanName(name)
					}
				}
			}
			// Or on next line
			if i+1 < len(lines) {
				name := strings.TrimSpace(lines[i+1])
				if len(name) > 3 {
					return cleanName(name)
				}
			}
		}
	}

	return ""
}

func cleanName(name string) string {
	// Remove common OCR artifacts and noise
	name = strings.TrimSpace(name)
	// Remove dates, numbers at start
	words := strings.Fields(name)
	var cleanWords []string
	for _, word := range words {
		// Skip if it looks like a date or pure number
		if len(word) > 1 && !strings.Contains(word, "202") && !strings.Contains(word, ".") {
			cleanWords = append(cleanWords, word)
		}
	}
	return strings.Join(cleanWords, " ")
}

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
		log.Fatal("Usage: go run extract_names.go <pdf_path>")
	}

	pdfPath := os.Args[1]
	doc, err := fitz.New(pdfPath)
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	fmt.Printf("PDF má %d strán\n\n", doc.NumPage())
	fmt.Println("Strana | QR ID | Rozpoznané meno")
	fmt.Println("-------|-------|------------------")

	for pageNum := 0; pageNum < doc.NumPage(); pageNum++ {
		img, err := doc.Image(pageNum)
		if err != nil {
			fmt.Printf("%-6d | ERROR | Chyba pri načítaní\n", pageNum+1)
			continue
		}

		// Convert to gocv.Mat
		mat, err := gocv.ImageToMatRGB(img)
		if err != nil {
			fmt.Printf("%-6d | ERROR | Chyba konverzie\n", pageNum+1)
			continue
		}

		// Read QR code
		qrText := readQRFromMat(mat)

		// Extract name via OCR
		name := extractNameOCR(mat)

		mat.Close()

		if qrText != "" {
			fmt.Printf("%-6d | %-5s | %s\n", pageNum+1, qrText, name)
		} else {
			fmt.Printf("%-6d | N/A   | %s\n", pageNum+1, name)
		}
	}
}
