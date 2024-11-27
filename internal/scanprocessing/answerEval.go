package scanprocessing

import (
	"ScanEvalApp/internal/ocr"
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Evaluate answers
func EvaluateAnswers(mat *gocv.Mat, numberOfQuestionsPerPage int) {
	croppedMat := CropMatAnswersOnly(mat)
	questionNumber := -1
	//contours := FindContours(MatToGrayscale(croppedMat))
	for i := 0; i < numberOfQuestionsPerPage; i++ {
		if questionNumber == -1 {
			questionNumber = GetQuestionNumber(&croppedMat, i, numberOfQuestionsPerPage)
		} else {
			questionNumber++
		}
		rectCheckboxes := image.Rectangle{Min: image.Point{PADDING + (croppedMat.Cols() / (NUMBER_OF_CHOICES + 1)), PADDING + (i * croppedMat.Rows() / numberOfQuestionsPerPage)}, Max: image.Point{croppedMat.Cols() - PADDING, ((i + 1) * croppedMat.Rows() / numberOfQuestionsPerPage) - PADDING}}
		questionMat := croppedMat.Region(rectCheckboxes)
		ShowMat(questionMat)
		//DrawRectangle(&croppedMat, rect)
	}
	*mat = croppedMat
}

// Crop image to contain only answers
func CropMatAnswersOnly(mat *gocv.Mat) gocv.Mat {
	rect := FindBorderRectangle(mat)
	rectSmaller := image.Rectangle{Min: image.Point{rect.Min.X + PADDING, rect.Min.Y + PADDING}, Max: image.Point{rect.Max.X - PADDING, rect.Max.Y - PADDING}}
	croppedMat := mat.Region(rectSmaller)
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

func GetQuestionNumber(mat *gocv.Mat, i int, numberOfQuestionsPerPage int) int {
	rect := image.Rectangle{Min: image.Point{PADDING, PADDING + (i * mat.Rows() / numberOfQuestionsPerPage)}, Max: image.Point{(mat.Cols() / (NUMBER_OF_CHOICES + 1)) - PADDING, ((i + 1) * mat.Rows() / numberOfQuestionsPerPage) - PADDING}}
	questionMat := mat.Region(rect)
	defer questionMat.Close()
	ShowMat(questionMat)
	SaveMat("", questionMat)
	dt := ocr.OcrImage(TEMP_IMAGE_PATH, ocr.PSM_SINGLE_LINE)
	var num int
	_, err := fmt.Sscan(dt, &num)
	if err != nil {
		fmt.Println("Conversion error:", err)
		return -1
	}
	return num
}
