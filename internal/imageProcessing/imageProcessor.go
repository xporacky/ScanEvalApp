package imageprocessing

import (
	"fmt"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

func Pdf2Img(scanPath string, outputPath string) {
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
			err = os.MkdirAll(outputPath, 0755)
			if err != nil {
				panic(err)
			}

			f, err := os.Create(filepath.Join(outputPath+"/", fmt.Sprintf("%s-image-%05d.jpg", folder, n)))
			if err != nil {
				panic(err)
			}

			err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
			if err != nil {
				panic(err)
			}

			f.Close()

		}
	}
}
