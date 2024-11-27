package scanprocessing

import (
	"ScanEvalApp/internal/ocr"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

// Process PDF
func ProcessPDF(scanPath string, outputPath string) {
	var files []string

	err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".pdf" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		doc, err := fitz.New(file)
		if err != nil {
			panic(err)
		}
		folder := strings.TrimSuffix(path.Base(file), filepath.Ext(path.Base(file)))

		// Extract pages as images
		for n := 0; n < doc.NumPage(); n++ {
			img, err := doc.Image(n)
			if err != nil {
				panic(err)
			}
			mat := ImageToMat(img)
			mat = MatToGrayscale(mat)
			mat = FixImageRotation(mat)
			qrText := readQR(&mat)
			if qrText == "" {
				// TODO
				fmt.Println("QR code was not found trying to found student code from OCR")
			}
			fmt.Println(qrText)
			path := filepath.Join("./"+outputPath+"/", fmt.Sprintf("%s-image-%05d.png", folder, n)) //na testovanie zatial takto
			EvaluateAnswers(&mat, NUMBER_OF_QUESTIONS_PER_PAGE)
			//path := filepath.Join("./"+outputPath+"/temp-image.png")
			SaveMat(path, mat)
			textInImg := ocr.OcrImage(path)
			fmt.Println(textInImg)
			//fmt.Println(path)
			ShowMat(mat)
			//return
		}
	}
}
