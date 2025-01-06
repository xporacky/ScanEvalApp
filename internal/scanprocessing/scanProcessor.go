package scanprocessing

import (
	"fmt"
	"os"
	"path/filepath"

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
		//folder := strings.TrimSuffix(path.Base(file), filepath.Ext(path.Base(file)))

		// Extract pages as images
		for n := 0; n < doc.NumPage(); n++ {
			ProcessPage(doc, n)
		}
	}
}

func ProcessPage(doc *fitz.Document, n int) {
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
		return
	}
	fmt.Println("ID STUDENTA: ", qrText)
	//path := filepath.Join("./"+outputPath+"/", fmt.Sprintf("%s-image-%05d.png", folder, n)) //na testovanie zatial takto
	EvaluateAnswers(&mat, 50)
	//SaveMat("", mat)
	defer mat.Close()
	//ShowMat(mat)
	//return
}
