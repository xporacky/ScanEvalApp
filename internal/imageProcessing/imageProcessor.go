package imageprocessing

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"gocv.io/x/gocv"
	"golang.org/x/image/draw"
)

const DPI = 300

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
			//img = increaseDPI(img, DPI)
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

// Increases image quality by increasing dpi bud does not change dpi in metadata
func increaseDPI(img *image.RGBA, dpi int) *image.RGBA {
	newWidth := int(float64(img.Bounds().Dx()) * float64(dpi) / 96.0)
	newHeight := int(float64(img.Bounds().Dy()) * float64(dpi) / 96.0)

	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(newImg, newImg.Rect, img, img.Bounds(), draw.Over, nil)
	return newImg
}

func imageToMat(imgRGBA *image.RGBA) gocv.Mat {
	width := imgRGBA.Bounds().Dx()
	height := imgRGBA.Bounds().Dy()
	buffer := new(bytes.Buffer)
	err := jpeg.Encode(buffer, imgRGBA, &jpeg.Options{Quality: jpeg.DefaultQuality})
	if err != nil {
		panic(err)
	}
	img, err := gocv.NewMatFromBytes(height, width, gocv.MatTypeCV8UC3, buffer.Bytes())
	if err != nil {
		panic(err)
	}
	return img
}
