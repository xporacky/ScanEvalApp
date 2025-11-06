package scanprocessing

import (
	"image"
	"math"

	"ScanEvalApp/internal/logging"

	"gocv.io/x/gocv"
)

// FindBorderRotatedRectangle detects the border of an answer sheet in the given image and returns
// the bounding rotated rectangle that encloses the sheet.
//
// The function finds contours in the image and approximates the shapes. If a quadrilateral (rectangle)
// with a sufficiently large area is found, it calculates and returns the minimum bounding rotated rectangle
// that encloses the answer sheet. This is useful for extracting the region of interest (ROI) containing the
// answer sheet, especially when it is rotated or skewed.
//
// Parameters:
//   - mat: A gocv.Mat representing the input image containing the answer sheet.
//
// Returns:
//   - gocv.RotatedRect: The rotated rectangle that bounds the detected answer sheet. If no valid rectangle
//     is found, it returns an empty `gocv.RotatedRect{}`.
//
// Notes:
//   - The function assumes that the answer sheet is large enough to be detected.
func FindBorderRotatedRectangle(mat gocv.Mat) gocv.RotatedRect {
	logger := logging.GetLogger()

	contours := FindContours(mat)
	// Find rectangle
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		approx := gocv.ApproxPolyDP(c, 0.01*gocv.ArcLength(c, true), true)
		if approx.Size() == 4 && gocv.ContourArea(approx) > BORDER_RECTANGLE_AREA_SIZE {
			rect := gocv.MinAreaRect(approx)
			//DrawRotatedRectangle(mat, rect)
			return rect
		}
	}
	logger.Info("Ohraničujúci obdĺžnik nebol nájdený")
	return gocv.RotatedRect{}
}

// FixImageRotation rotates the input image to correct its orientation based on detected skew or rotation.
//
// The function first detects the bounding rotated rectangle of the answer sheet using `FindBorderRotatedRectangle`
// and calculates the angle needed to correct the image's orientation. If the image is upside down, the angle is adjusted
// by 180 degrees. The image is then rotated around its center using the calculated angle to ensure the sheet is oriented
// properly. This is useful for correcting rotated or skewed answer sheets before further processing.
//
// Parameters:
//   - mat: A gocv.Mat representing the input image that may need rotation.
//
// Returns:
//   - gocv.Mat: The rotated image that has been corrected to have the proper orientation.
//
// Notes:
//   - The function uses the center of the image for rotation and assumes the presence of a detectable border rectangle.
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

// CheckUpsideDown checks if the input image is upside down by comparing the contours of the upper and lower halves.
//
// The function divides the image into two halves: the upper half and the lower half. It then counts the number of contours
// in each half using `FindContours`. If the lower half has more contours than the upper half, the image is likely upside down,
// as the lower part typically contains more significant features (such as text or form fields). This is used to determine the
// orientation of the image and whether it needs to be rotated for proper alignment.
//
// Parameters:
//   - mat: A gocv.Mat representing the input image to check for orientation.
//
// Returns:
//   - bool: Returns `true` if the image is upside down (i.e., the lower half has more contours than the upper half),
//     and `false` otherwise.
//
// Notes:
//   - The function assumes that the top half of the image typically contains more features (e.g., header).
func CheckUpsideDown(mat gocv.Mat) bool {
	upperPart := mat.Region(image.Rect(0, 0, mat.Cols(), mat.Rows()/2))
	lowerPart := mat.Region(image.Rect(0, mat.Rows()/2, mat.Cols(), mat.Rows()))
	return FindContours(lowerPart).Size() > FindContours(upperPart).Size()
}
