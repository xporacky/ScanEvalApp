package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"fmt"
	"image"
	"image/color"
	"math"

	"gocv.io/x/gocv"
	"golang.org/x/image/draw"
)

func CalculatePointsDistance(p1, p2 image.Point) float64 {
	return math.Sqrt(math.Pow(float64(p1.X-p2.X), 2) + math.Pow(float64(p1.Y-p2.Y), 2))
}

func FindContours(mat gocv.Mat) gocv.PointsVector {
	// Use Canny edge detection
	canny := gocv.NewMat()
	defer canny.Close()
	gocv.Canny(mat, &canny, 100, 200)

	// Use morphological closing
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Dilate(canny, &canny, kernel)
	gocv.Erode(canny, &canny, kernel)

	// Find contours
	contours := gocv.FindContours(canny, gocv.RetrievalExternal, gocv.ChainApproxNone)
	fmt.Println("Found", contours.Size(), "contours")
	return contours
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

// Convert image to grayscale
func MatToGrayscale(mat gocv.Mat) gocv.Mat {
	gray := gocv.NewMat()
	gocv.CvtColor(mat, &gray, gocv.ColorBGRToGray)
	return gray
}

// Save image in gocv.Mat to png file
func SaveMat(path string, mat gocv.Mat) {
	if path == "" {
		path = TEMP_IMAGE_PATH
	}
	err := files.DeleteFile(path)
	if err != nil {
		panic(err)
	}
	gocv.IMWrite(path, mat)
}

func readQR(mat *gocv.Mat) string {
	qrDetector := gocv.NewQRCodeDetector()
	points := gocv.NewMat()
	defer points.Close()
	qrCode := gocv.NewMat()
	defer qrCode.Close()
	text := qrDetector.DetectAndDecode(*mat, &points, &qrCode)
	return text
}

// Draw red rotated rectangle on image
func DrawRotatedRectangle(mat *gocv.Mat, rect gocv.RotatedRect) {
	color := color.RGBA{255, 0, 0, 255}
	rectPoints := rect.Points
	for i := 0; i < 4; i++ {
		gocv.Line(mat, rectPoints[i], rectPoints[(i+1)%4], color, 10)
	}
}

// Draw red rectangle on image
func DrawRectangle(mat *gocv.Mat, rect image.Rectangle) {
	color := color.RGBA{255, 0, 0, 255}
	gocv.Rectangle(mat, rect, color, 10)
}

func drawCountours(mat *gocv.Mat, contours gocv.PointsVector) {
	for i := 0; i < contours.Size(); i++ {
		gocv.DrawContours(mat, contours, i, color.RGBA{255, 0, 0, 255}, 10)
	}
}
