package scanprocessing

import (
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/ocr"
	"image"

	"ScanEvalApp/internal/logging"
	"log/slog"

	"gocv.io/x/gocv"
)

// EvaluateAnswers processes a scanned answer sheet image and extracts the student's answers.
//
// It takes a pointer to a gocv.Mat representing the scanned sheet and the total number of questions expected.
// The function crops the input image to focus on the answers section and iterates over detected answer regions.
// It attempts to determine the starting question number by reading any visible question numbers on the page.
// Once the starting number is found, it continues incrementally until all answers are extracted or the
// total number of questions is reached.
//
// If no question number is detected on the page, the function logs an error and returns -1 with a nil slice.
//
// Parameters:
//   - mat: A pointer to the original gocv.Mat image. This will be modified in-place to the cropped version.
//   - numberOfQuestions: Total number of questions expected across the form.
//
// Returns:
//   - int: The index of the last question found (or -1 if none were found).
//   - []rune: A slice containing the student's selected answers as runes (e.g., 'A', 'B', 'C', etc.).
func EvaluateAnswers(mat *gocv.Mat, numberOfQuestions int) (int, []rune) {
	logger := logging.GetLogger()
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
		return -1, nil
	}
	return questionNumber - 1, studentAnswers
}

// CropMatAnswersOnly extracts the region of the image that contains only the answers.
//
// It finds the bounding rectangle that likely surrounds the answer area using the provided constants,
// then shrinks it slightly using padding to exclude borders or noise.
// The function returns a new gocv.Mat cropped to this inner region.
//
// Parameters:
//   - mat: A pointer to a gocv.Mat representing the original scanned sheet.
//
// Returns:
//   - gocv.Mat: A new Mat representing the cropped image region containing only the answers.
func CropMatAnswersOnly(mat *gocv.Mat) gocv.Mat {
	rect := FindRectangle(mat, BORDER_RECTANGLE_AREA_SIZE, -1)
	rectSmaller := image.Rectangle{Min: image.Point{rect.Min.X + PADDING, rect.Min.Y + PADDING}, Max: image.Point{rect.Max.X - PADDING, rect.Max.Y - PADDING}}
	croppedMat := mat.Region(rectSmaller)
	return croppedMat
}

// FindRectangle detects and returns the bounding rectangle of a contour in the image.
//
// It processes the input image to find contours and approximates their shapes.
// If a contour has at least four points and its area is within the specified range,
// its bounding rectangle is returned. The function prioritizes the first valid match.
//
// If no valid rectangle is found, it logs a warning and returns an empty rectangle.
//
// Parameters:
//   - mat: A pointer to a gocv.Mat representing the source image.
//   - minAreaSize: The minimum area required for a contour to be considered.
//   - maxAreaSize: The maximum area allowed for a contour. If set to -1, no upper limit is applied.
//
// Returns:
//   - image.Rectangle: The bounding rectangle of the detected contour, or an empty rectangle if none found.
func FindRectangle(mat *gocv.Mat, minAreaSize float64, maxAreaSize float64) image.Rectangle {
	errorLogger := logging.GetErrorLogger()
	contours := FindContours(*mat)
	// Find rectangle
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		approx := gocv.ApproxPolyDP(c, 0.01*gocv.ArcLength(c, true), true)
		if approx.Size() >= 4 && gocv.ContourArea(approx) > minAreaSize {
			if maxAreaSize != -1 && gocv.ContourArea(approx) > maxAreaSize {
				continue
			}
			rect := gocv.BoundingRect(approx)
			return rect
		}
	}
	errorLogger.Warn("Nezistený obvod v matici", "error", "No valid rectangle found")
	return image.Rectangle{image.Pt(0, 0), image.Pt(0, 0)}
}

// GetQuestionNumber attempts to extract the question number from a specific region of the image using OCR.
//
// The function calculates a rectangular region within the image where the question number is expected,
// based on the index of the question and predefined constants. It crops that region, saves it as a temporary
// image, and uses OCR to extract the number. After processing, the temporary image is deleted.
//
// Parameters:
//   - mat: A pointer to a gocv.Mat representing the cropped answer section.
//   - i: The index of the question within the current page.
//
// Returns:
//   - int: The extracted question number. If OCR fails, it returns zero (default int value).
func GetQuestionNumber(mat *gocv.Mat, i int) int {
	errorLogger := logging.GetErrorLogger()
	rect := image.Rectangle{Min: image.Point{PADDING, PADDING + (i * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE)}, Max: image.Point{(mat.Cols() / (NUMBER_OF_CHOICES + 1)) - PADDING, ((i + 1) * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE) - PADDING}}
	questionMat := mat.Region(rect)
	defer questionMat.Close()
	SaveMat(TEMP_IMAGE_PATH, questionMat)
	questionNum, err := ocr.ExtractQuestionNumber(TEMP_IMAGE_PATH)
	files.DeleteFile(TEMP_IMAGE_PATH)

	if err != nil {
		errorLogger.Error("Chyba pri extrakcii čísla otázky", slog.Int("questionIndex", i), slog.String("error", err.Error()))
	}

	return questionNum
}

// GetAnswer evaluates a single question's answer by analyzing the corresponding row of checkboxes.
//
// For a given question index `i`, the function scans through all possible answer choices (e.g., A–D),
// determines the checkbox area for each choice, and analyzes its content to detect a marked answer.
// It first checks if a rectangular area (checkbox) is present. If found, it examines the mean intensity
// of the inner region to decide whether the box is marked with an "X" or left empty.
//
// The function uses internal states to detect ambiguous markings (e.g., multiple selections) and
// returns 'x' in such cases to indicate an invalid or unclear answer.
//
// Parameters:
//   - mat: A pointer to a gocv.Mat representing the cropped image of answers.
//   - i: The index of the question within the current page (0-based).
//
// Returns:
//   - rune: The selected answer (e.g., 'a', 'b', 'c', etc.). Returns 'x' if no valid or multiple answers are detected.
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
