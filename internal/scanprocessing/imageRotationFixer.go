package scanprocessing

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
)

// Finds border rectangle of asnwer sheet
func FindBorderRotatedRectangle(mat gocv.Mat) gocv.RotatedRect {
	contours := FindContours(mat)
	// Find rectangle
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		approx := gocv.ApproxPolyDP(c, 0.01*gocv.ArcLength(c, true), true)
		if approx.Size() == 4 && gocv.ContourArea(approx) > 1000000 {
			fmt.Println(gocv.ContourArea(approx))
			rect := gocv.MinAreaRect(approx)
			//DrawRotatedRectangle(mat, rect)
			return rect
		}
	}
	return gocv.RotatedRect{}
}

// Rotate image by center
func FixImageRotation(mat gocv.Mat) gocv.Mat {
	rect := FindBorderRotatedRectangle(mat)
	// Rotate image
	angle := rect.Angle - 90
	if math.Abs(angle) > 45 {
		angle += 90
	}
	if CheckUpsideDown(mat) {
		angle += 180
	}
	rotationMatrix := gocv.GetRotationMatrix2D(image.Pt(mat.Cols()/2, mat.Rows()/2), angle, 1)
	rotated := gocv.NewMat()
	size := mat.Size()
	gocv.WarpAffine(mat, &rotated, rotationMatrix, image.Pt(size[1], size[0]))
	return rotated
}

// Check if image is upside down
func CheckUpsideDown(mat gocv.Mat) bool {
	upperPart := mat.Region(image.Rect(0, 0, mat.Cols(), mat.Rows()/2))
	lowerPart := mat.Region(image.Rect(0, mat.Rows()/2, mat.Cols(), mat.Rows()))
	//fmt.Println("Upper part size:", FindContours(upperPart).Size(), "Lower part size:", FindContours(lowerPart).Size())
	return FindContours(lowerPart).Size() > FindContours(upperPart).Size()
}
