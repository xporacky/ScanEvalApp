package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"gocv.io/x/gocv"
	"golang.org/x/image/draw"
)

const DPI = 300

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
			mat = ProccessMat(mat)
			path := filepath.Join("./"+outputPath+"/", fmt.Sprintf("%s-image-%05d.png", folder, n)) //na testovanie zatial takto
			//path := filepath.Join("./"+outputPath+"/temp-image.png")
			SaveMat(path, mat)
			textInImg := ocr.OcrImage(path)
			fmt.Println(textInImg)
			ShowMat(mat)
			//return
		}
	}
}

// Increases image quality by increasing dpi bud does not change dpi in metadata
func IncreaseDPI(img *image.RGBA, dpi int) *image.RGBA {
	newWidth := int(float64(img.Bounds().Dx()) * float64(dpi) / 96.0)
	newHeight := int(float64(img.Bounds().Dy()) * float64(dpi) / 96.0)

	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(newImg, newImg.Rect, img, img.Bounds(), draw.Over, nil)
	return newImg
}

// Converts image to gocv.Mat
func ImageToMat(imgRGBA *image.RGBA) gocv.Mat {
	bounds := imgRGBA.Bounds()
	x := bounds.Dx()
	y := bounds.Dy()
	bytes := make([]byte, 0, x*y)
	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			r, g, b, _ := imgRGBA.At(i, j).RGBA()
			bytes = append(bytes, byte(b>>8))
			bytes = append(bytes, byte(g>>8))
			bytes = append(bytes, byte(r>>8))
		}
	}

	mat, err := gocv.NewMatFromBytes(y, x, gocv.MatTypeCV8UC3, bytes)
	//img, err := gocv.NewMatFromBytes(height, width, gocv.MatTypeCV8UC3, buffer.Bytes())
	if err != nil {
		panic(err)
	}
	return mat
}

// Show image in window
func ShowMat(mat gocv.Mat) {
	window := gocv.NewWindow("Image")
	defer window.Close()
	window.ResizeWindow(1100, 1400)
	window.IMShow(mat)
	window.WaitKey(0)
}

// Process image with OpenCV
func ProccessMat(mat gocv.Mat) gocv.Mat {
	// Convert image to grayscale
	gray := gocv.NewMat()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)

	// Use Canny edge detection
	canny := gocv.NewMat()
	defer canny.Close()
	gocv.Canny(gray, &canny, 100, 200)

	// Use morphological closing
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(canny, &canny, kernel)
	gocv.Erode(canny, &canny, kernel)

	// Find contours
	contours := gocv.FindContours(canny, gocv.RetrievalExternal, gocv.ChainApproxNone)
	fmt.Println("Found", contours.Size(), "contours")

	// Draw contours
	for i := 0; i < contours.Size(); i++ {
		//c := contours.At(i)
		color := color.RGBA{255, 0, 0, 255}
		gocv.DrawContours(&gray, contours, i, color, 1)
	}

	// Show the result
	return gray
}

// Save image in gocv.Mat to png file
func SaveMat(path string, mat gocv.Mat) {
	err := files.DeleteFile(path)
	if err != nil {
		panic(err)
	}
	gocv.IMWrite(path, mat)
}
