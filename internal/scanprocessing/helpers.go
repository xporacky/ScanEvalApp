package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"fmt"
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

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
	//fmt.Println("Found", contours.Size(), "contours")
	return contours
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
	fmt.Println("Succesfully saved file: ", TEMP_IMAGE_PATH)
}

func ReadQR(mat *gocv.Mat) string {
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

func DrawCountours(mat *gocv.Mat, contours gocv.PointsVector) {
	for i := 0; i < contours.Size(); i++ {
		gocv.DrawContours(mat, contours, i, color.RGBA{255, 0, 0, 255}, 5)
	}
}

func GetStudentID(mat *gocv.Mat) (int, error) {
	qrText := ReadQR(mat)
	if qrText != "" {
		var id int
		_, err := fmt.Sscan(qrText, &id)
		if err != nil {
			return 0, fmt.Errorf("failed to convert ID to integer: %v", err)
		}
		return id, nil
	}
	rect := image.Rectangle{Min: image.Point{PADDING, PADDING}, Max: image.Point{mat.Cols() - PADDING, (mat.Rows() / 4) - PADDING}}
	headerMat := mat.Region(rect)
	defer headerMat.Close()
	SaveMat(TEMP_HEADER_IMAGE_PATH, headerMat)
	id, err := ocr.ExtractID(TEMP_HEADER_IMAGE_PATH)
	files.DeleteFile(TEMP_IMAGE_PATH)
	if err != nil {
		return 0, err
	}
	return id, nil
}
