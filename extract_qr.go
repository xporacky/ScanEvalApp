package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

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
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run extract_qr.go <pdf_file>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	fmt.Printf("Analyzing PDF: %s\n\n", pdfPath)
	fmt.Println("============================================================")

	// Open PDF
	doc, err := fitz.New(pdfPath)
	if err != nil {
		log.Fatalf("Error opening PDF: %v", err)
	}
	defer doc.Close()

	pageCount := doc.NumPage()
	fmt.Printf("Total pages: %d\n\n", pageCount)

	// Map QR code -> pages
	qrMapping := make(map[string][]int)

	// Process each page
	for pageNum := 0; pageNum < pageCount; pageNum++ {
		// Extract page as image
		img, err := doc.Image(pageNum)
		if err != nil {
			fmt.Printf("Page %3d: Error extracting image: %v\n", pageNum+1, err)
			continue
		}

		// Convert to gocv.Mat
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// Convert image.Image to RGB bytes
		rgbBytes := make([]byte, 0, width*height*3)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				rgbBytes = append(rgbBytes, byte(b>>8), byte(g>>8), byte(r>>8))
			}
		}

		mat, err := gocv.NewMatFromBytes(height, width, gocv.MatTypeCV8UC3, rgbBytes)
		if err != nil {
			fmt.Printf("Page %3d: Error converting to Mat: %v\n", pageNum+1, err)
			continue
		}

		// Read QR code
		qrText := readQRFromMat(mat)
		mat.Close()

		if qrText != "" {
			if _, exists := qrMapping[qrText]; !exists {
				qrMapping[qrText] = []int{}
			}
			qrMapping[qrText] = append(qrMapping[qrText], pageNum+1)
			fmt.Printf("Page %3d: QR Code = %s\n", pageNum+1, qrText)
		} else {
			fmt.Printf("Page %3d: No QR code found\n", pageNum+1)
		}
	}

	fmt.Println("\n============================================================")
	fmt.Println("\nSummary - QR Code to Pages mapping:")
	fmt.Println("------------------------------------------------------------")

	if len(qrMapping) == 0 {
		fmt.Println("No QR codes found in the PDF")
	} else {
		// Sort QR codes numerically
		qrCodes := make([]string, 0, len(qrMapping))
		for qr := range qrMapping {
			qrCodes = append(qrCodes, qr)
		}
		sort.Slice(qrCodes, func(i, j int) bool {
			numI, errI := strconv.Atoi(qrCodes[i])
			numJ, errJ := strconv.Atoi(qrCodes[j])
			if errI == nil && errJ == nil {
				return numI < numJ
			}
			return qrCodes[i] < qrCodes[j]
		})

		for _, qr := range qrCodes {
			pages := qrMapping[qr]
			pagesStr := ""
			for i, p := range pages {
				if i > 0 {
					pagesStr += ", "
				}
				pagesStr += strconv.Itoa(p)
			}
			fmt.Printf("QR Code %4s -> Pages: %s\n", qr, pagesStr)
		}
	}
}
