package scanprocessing

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

func cropMatAnswersOnly(mat *gocv.Mat) gocv.Mat {
	croppedMat := mat.Region(FindBorderRectangle(mat))
	return croppedMat
}

// Finds border rectangle of asnwer sheet
func FindBorderRectangle(mat *gocv.Mat) image.Rectangle {
	contours := FindContours(*mat)
	// Find rectangle
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		approx := gocv.ApproxPolyDP(c, 0.01*gocv.ArcLength(c, true), true)
		if approx.Size() == 4 && gocv.ContourArea(approx) > 1000000 {
			fmt.Println(gocv.ContourArea(approx))
			rect := gocv.BoundingRect(approx)
			//DrawRectangle(mat, rect)
			return rect
		}
	}
	return image.Rectangle{}
}
