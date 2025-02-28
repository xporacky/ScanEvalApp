package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"image"

	"ScanEvalApp/internal/logging"
	"log/slog"

	"gocv.io/x/gocv"
)

// Evaluate answers
func EvaluateAnswers(mat *gocv.Mat, numberOfQuestions int) (int, []rune) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	var studentAnswers []rune
	croppedMat := CropMatAnswersOnly(mat)
	questionNumber := 0
	for i := 0; i < NUMBER_OF_QUESTIONS_PER_PAGE; i++ {
		studentAnswers = append(studentAnswers, GetAnswer(&croppedMat, i))
		// if we dont have question number yet try to find it
		if questionNumber == 0 {
			questionNumber = GetQuestionNumber(&croppedMat, i)
			continue
		}
		questionNumber++
		if questionNumber > numberOfQuestions {
			logger.Info("Všetky otázky boli nájdené")
			break
		}

	}
	*mat = croppedMat
	// if we didnt find question number in whole page
	if questionNumber == -1 {
		//TODO nejaky fail safe
		errorLogger.Error("Neboli nájdené žiadne čísla otázok", "error", "No question number found")
		return -1, nil
	}
	return questionNumber - 1, studentAnswers
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
	errorLogger := logging.GetErrorLogger()
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
	errorLogger.Warn("Nezistený obvod v matici", "error", "No valid rectangle found")
	return image.Rectangle{image.Pt(0, 0), image.Pt(0, 0)}
}

// Gets Number of question using ocr
func GetQuestionNumber(mat *gocv.Mat, i int) int {
	errorLogger := logging.GetErrorLogger()
	rect := image.Rectangle{Min: image.Point{PADDING, PADDING + (i * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE)}, Max: image.Point{(mat.Cols() / (NUMBER_OF_CHOICES + 1)) - PADDING, ((i + 1) * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE) - PADDING}}
	questionMat := mat.Region(rect)
	defer questionMat.Close()
	//ShowMat(questionMat)
	SaveMat(TEMP_IMAGE_PATH, questionMat)
	questionNum, err := ocr.ExtractQuestionNumber(TEMP_IMAGE_PATH)
	files.DeleteFile(TEMP_IMAGE_PATH)

	if err != nil {
		errorLogger.Error("Chyba pri extrakcii čísla otázky", slog.Int("questionIndex", i), slog.String("error", err.Error()))
	}

	return questionNum
}

// Evaluate one questions returns answer to this question
func GetAnswer(mat *gocv.Mat, i int) rune {
	answer := rune('x')
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
				return rune('x')
			}
			answer = rune('a' + (j - 1))
			state = StateCircleFound
			continue
		}
		checkboxWithoutBorder := image.Rectangle{Min: image.Point{rect.Min.X + CHECKBOX_PADDING, rect.Min.Y + CHECKBOX_PADDING}, Max: image.Point{rect.Max.X - CHECKBOX_PADDING, rect.Max.Y - CHECKBOX_PADDING}}
		rectMat := checkboxMat.Region(checkboxWithoutBorder)
		meanIntensity := rectMat.Mean()
		if meanIntensity.Val1 < MEAN_INTENSITY_X_HIGHEST && meanIntensity.Val1 > MEAN_INTENSITY_X_LOWEST {
			if state == StateEmpty {
				answer = rune('a' + (j - 1))
				state = StateXFound
				continue
			} else if state == StateXFound {
				answer = rune('x')
			}
		}
		//fmt.Println(meanIntensity.Val1)
		defer checkboxMat.Close()
		defer rectMat.Close()
	}
	return answer
}
