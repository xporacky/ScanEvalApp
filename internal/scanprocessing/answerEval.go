package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"fmt"
	"image"
	"log"

	"gocv.io/x/gocv"
)

// Evaluate answers
func EvaluateAnswers(mat *gocv.Mat, numberOfQuestions int) {
	var unknownQuestionsAnswers []string
	croppedMat := CropMatAnswersOnly(mat)
	questionNumber := -1
	for i := 0; i < NUMBER_OF_QUESTIONS_PER_PAGE; i++ {
		answer := GetAnswer(&croppedMat, i)
		// if we dont have question number yet try to find it
		if questionNumber == -1 {
			questionNumber = GetQuestionNumber(&croppedMat, i)
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
		log.Println(questionNumber, " | ", answer)
		questionNumber++
		if questionNumber > numberOfQuestions {
			fmt.Println("All questions found")
			break
		}

	}
	*mat = croppedMat
	// if we didnt find question number in whole page
	if questionNumber == -1 {
		//TODO nejaky fail safe
		fmt.Println("No question number found")
	}
}

// Crop image to contain only answers
func CropMatAnswersOnly(mat *gocv.Mat) gocv.Mat {
	rect := FindRectangle(mat, BORDER_RECTANGLE_AREA_SIZE, -1)
	rectSmaller := image.Rectangle{Min: image.Point{rect.Min.X + PADDING, rect.Min.Y + PADDING}, Max: image.Point{rect.Max.X - PADDING, rect.Max.Y - PADDING}}
	croppedMat := mat.Region(rectSmaller)
	return croppedMat
}

// Finds rectangle on mat
func FindRectangle(mat *gocv.Mat, minAreaSize float64, maxAreaSize float64) image.Rectangle {
	contours := FindContours(*mat)
	// Find rectangle
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		approx := gocv.ApproxPolyDP(c, 0.01*gocv.ArcLength(c, true), true)
		//fmt.Println(gocv.ContourArea(approx), approx.Size())
		if approx.Size() >= 4 && gocv.ContourArea(approx) > minAreaSize {
			if maxAreaSize != -1 && gocv.ContourArea(approx) > maxAreaSize {
				continue
			}
			rect := gocv.BoundingRect(approx)
			//DrawRectangle(mat, rect)
			return rect
		}
	}
	return image.Rectangle{image.Pt(0, 0), image.Pt(0, 0)}
}

func GetQuestionNumber(mat *gocv.Mat, i int) int {
	rect := image.Rectangle{Min: image.Point{PADDING, PADDING + (i * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE)}, Max: image.Point{(mat.Cols() / (NUMBER_OF_CHOICES + 1)) - PADDING, ((i + 1) * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE) - PADDING}}
	questionMat := mat.Region(rect)
	defer questionMat.Close()
	//ShowMat(questionMat)
	SaveMat(TEMP_IMAGE_PATH, questionMat)
	dt := ocr.OcrImage(TEMP_IMAGE_PATH, ocr.PSM_SINGLE_LINE)
	var num int
	_, err := fmt.Sscan(dt, &num)
	files.DeleteFile(TEMP_IMAGE_PATH)
	if err != nil {
		fmt.Println("Conversion error:", err)
		return -1
	}
	fmt.Println("Question number:", num)
	return num
}

func GetAnswer(mat *gocv.Mat, i int) string {
	answer := ""
	state := StateEmpty
	for j := 1; j <= NUMBER_OF_CHOICES; j++ {
		padding := CHECKBOX_AREA_PADDING
		if i == 0 || i == NUMBER_OF_QUESTIONS_PER_PAGE-1 {
			padding = 0
		}
		checkbox := image.Rectangle{Min: image.Point{(mat.Cols() / (NUMBER_OF_CHOICES + 1) * (j)), padding + (i * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE)}, Max: image.Point{(mat.Cols() / (NUMBER_OF_CHOICES + 1)) * (j + 1), ((i + 1) * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE) - padding}}
		checkboxMat := mat.Region(checkbox)
		rect := FindRectangle(&checkboxMat, ANSWER_SQUARE_MIN_AREA_SIZE, ANSWER_SQUARE_MAX_AREA_SIZE)
		if rect.Empty() {
			if state == StateCircleFound {
				return ""
			}
			answer = string(rune('a' + (j - 1)))
			state = StateCircleFound
			continue
		}
		checkboxWithoutBorder := image.Rectangle{Min: image.Point{rect.Min.X + CHECKBOX_PADDING, rect.Min.Y + CHECKBOX_PADDING}, Max: image.Point{rect.Max.X - CHECKBOX_PADDING, rect.Max.Y - CHECKBOX_PADDING}}
		rectMat := checkboxMat.Region(checkboxWithoutBorder)
		meanIntensity := rectMat.Mean()
		if meanIntensity.Val1 < MEAN_INTENSITY_X_HIGHEST && meanIntensity.Val1 > MEAN_INTENSITY_X_LOWEST {
			if state == StateEmpty {
				answer = string(rune('a' + (j - 1)))
				state = StateXFound
				continue
			} else if state == StateXFound {
				answer = ""
			}
		}
		//fmt.Println(meanIntensity.Val1)
		defer checkboxMat.Close()
		defer rectMat.Close()
	}
	return answer
}
