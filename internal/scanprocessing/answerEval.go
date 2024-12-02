package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Evaluate answers
func EvaluateAnswers(mat *gocv.Mat, numberOfQuestions int) {
	var unknownQuestionsAnswers []string
	croppedMat := CropMatAnswersOnly(mat)
	questionNumber := -1
	for i := 0; i < NUMBER_OF_QUESTIONS_PER_PAGE; i++ {
		answer := GetAnswer(&croppedMat, i, NUMBER_OF_QUESTIONS_PER_PAGE)
		// if we dont have question number yet try to find it
		if questionNumber == -1 {
			questionNumber = GetQuestionNumber(&croppedMat, i, NUMBER_OF_QUESTIONS_PER_PAGE)
			// if we didnt find question number yet add answer to unknown questions
			if questionNumber == -1 {
				unknownQuestionsAnswers = append(unknownQuestionsAnswers, answer)
			} else if questionNumber != -1 && unknownQuestionsAnswers != nil { // if we found question number and we have unknown questions answers assign them to question answers to student
				//TODO
				fmt.Println("Unknown questions answers:", unknownQuestionsAnswers)
			}

		} else {
			// TODO priradit odpoved k odpovediam studenta
		}
		questionNumber++
		if questionNumber > numberOfQuestions {
			fmt.Println("All questions found")
			break
		}

	}
	files.DeleteFile(TEMP_IMAGE_PATH)
	*mat = croppedMat
	// if we didnt find question number in whole page
	if questionNumber == -1 {
		//TODO nejaky fail safe
		fmt.Println("No question number found")
	}
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
	//ShowMat(questionMat)
	SaveMat("", questionMat)
	dt := ocr.OcrImage(TEMP_IMAGE_PATH, ocr.PSM_SINGLE_LINE)
	var num int
	_, err := fmt.Sscan(dt, &num)
	if err != nil {
		fmt.Println("Conversion error:", err)
		return -1
	}
	fmt.Println("Question number:", num)
	return num
}

func GetAnswer(mat *gocv.Mat, i int, numberOfQuestionsPerPage int) string {
	rectCheckboxes := image.Rectangle{Min: image.Point{PADDING + (mat.Cols() / (NUMBER_OF_CHOICES + 1)), PADDING + (i * mat.Rows() / numberOfQuestionsPerPage)}, Max: image.Point{mat.Cols() - PADDING, ((i + 1) * mat.Rows() / numberOfQuestionsPerPage) - PADDING}}
	questionMat := mat.Region(rectCheckboxes)
	drawCountours(&questionMat, FindContours(questionMat))
	//ShowMat(questionMat)
	defer questionMat.Close()
	return ""
}
