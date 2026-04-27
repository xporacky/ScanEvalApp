package scanprocessing

import (
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/logging"
	"ScanEvalApp/internal/ocr"
	"fmt"
	"image"

	"log/slog"

	"gocv.io/x/gocv"
)

var MEAN_INTENSITY_X_LOWEST float64
var MEAN_INTENSITY_X_HIGHEST float64

// checkboxPaddingForChoices returns the padding value to be used for checkbox inner area calculation
func answerSquareAreaBounds(choices int) (float64, float64) {
	if choices == 8 {
		return ANSWER_SQUARE_MIN_AREA_SIZE_8, ANSWER_SQUARE_MAX_AREA_SIZE_8
	}
	return ANSWER_SQUARE_MIN_AREA_SIZE, ANSWER_SQUARE_MAX_AREA_SIZE
}

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
func EvaluateAnswers(mat *gocv.Mat, numberOfQuestions int, numberOfChoices int) (int, []rune) {
	logger := logging.GetLogger()
	var studentAnswers []rune
	croppedMat, err := CropMatAnswersOnly(mat)
	if err != nil {
		logger.Info("Ohraničujúci obdĺžnik nebol nájdený", slog.String("error", err.Error()))
		return common.QUESTION_NUMBER_NOT_FOUND, nil
	}

	croppedUpperHeader, err := CropGroupFromHeader(mat)
	if err != nil {
		logger.Info("Ohraničujúci obdĺžnik nebol nájdený", slog.String("error", err.Error()))
		return common.QUESTION_NUMBER_NOT_FOUND, nil
	}
	defer croppedUpperHeader.Close()
	group := GetGroupCode(&croppedUpperHeader)

	croppedLowerHeader, err := CropSubGroupFromHeader(mat)
	if err != nil {
		logger.Info("Ohraničujúci obdĺžnik nebol nájdený", slog.String("error", err.Error()))
		return common.QUESTION_NUMBER_NOT_FOUND, nil
	}
	defer croppedLowerHeader.Close()
	subGroup := GetSubGroupCode(&croppedLowerHeader)

	if group == 'x' || subGroup == -1 {
		logger.Info("Nebola najdena skupina alebo podskupina", slog.Any("group", string(group)), slog.Int("subGroup", subGroup))
		return common.QUESTION_NUMBER_NOT_FOUND, nil
	}
	groupCode := fmt.Sprintf("%c%d", group, subGroup)
	fmt.Println("Skupina: ", groupCode)

	questionNumber := common.QUESTION_NUMBER_NOT_FOUND
	for i := 0; i < NUMBER_OF_QUESTIONS_PER_PAGE; i++ {
		studentAnswers = append(studentAnswers, GetAnswer(&croppedMat, i, numberOfChoices))
		// if we dont have question number yet try to find it
		if questionNumber == common.QUESTION_NUMBER_NOT_FOUND {
			questionNumber = GetQuestionNumber(&croppedMat, i, numberOfChoices)
			continue
		}
		questionNumber++
		if questionNumber >= numberOfQuestions {
			logger.Info("Všetky otázky boli nájdené")
			break
		}

	}
	*mat = croppedMat
	// if we didnt find question number in whole page
	if questionNumber == common.QUESTION_NUMBER_NOT_FOUND {
		return common.QUESTION_NUMBER_NOT_FOUND, nil
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
func CropMatAnswersOnly(mat *gocv.Mat) (gocv.Mat, error) {
	rect := FindRectangle(mat, BORDER_RECTANGLE_AREA_SIZE, -1)
	if rect.Empty() {
		return gocv.NewMat(), fmt.Errorf("ohranicujuci obdlznik nebol najdeny (prazdna strana?)")
	}
	rectSmaller := image.Rectangle{
		Min: image.Point{rect.Min.X + PADDING, rect.Min.Y + PADDING},
		Max: image.Point{rect.Max.X - PADDING, rect.Max.Y - PADDING},
	}
	if rectSmaller.Dx() <= 0 || rectSmaller.Dy() <= 0 {
		return gocv.NewMat(), fmt.Errorf("ohranicujuci obdlznik je prilis maly")
	}
	croppedMat := mat.Region(rectSmaller)
	return croppedMat, nil
}

// Crops a rectangle from the header
func CropGroupFromHeader(mat *gocv.Mat) (gocv.Mat, error) {
	rect := FindRectangle(mat, BORDER_RECTANGLE_AREA_SIZE, -1)
	if rect.Empty() {
		return gocv.NewMat(), fmt.Errorf("hlavny obdlznik nebol najdeny")
	}

	// Vyrezeme pas nad hlavnym obdlznikom.
	// Sirka ostane rovnaka ako hlavny obdlznik, vyska je len male pasmo nad nim.
	headerRect := image.Rectangle{
		Min: image.Point{
			X: rect.Min.X + GROUP_SIDE_PADDING,
			Y: rect.Min.Y - GROUP_HEADER_HEIGHT,
		},
		Max: image.Point{
			X: rect.Max.X - GROUP_SIDE_PADDING,
			Y: rect.Min.Y - GROUP_BOTTOM_PADDING,
		},
	}

	// Ochrana proti vybehnutiu mimo obrazka
	if headerRect.Min.X < 0 {
		headerRect.Min.X = 0
	}
	if headerRect.Min.Y < 0 {
		headerRect.Min.Y = 0
	}
	if headerRect.Max.X > mat.Cols() {
		headerRect.Max.X = mat.Cols()
	}
	if headerRect.Max.Y > mat.Rows() {
		headerRect.Max.Y = mat.Rows()
	}

	if headerRect.Dx() <= 0 || headerRect.Dy() <= 0 {
		return gocv.NewMat(), fmt.Errorf("oblast headera je prilis mala alebo mimo obrazka")
	}

	croppedMat := mat.Region(headerRect)
	SaveMat("./assets/group-crop.png", croppedMat)
	return croppedMat, nil
}

func CropSubGroupFromHeader(mat *gocv.Mat) (gocv.Mat, error) {
	rect := FindRectangle(mat, BORDER_RECTANGLE_AREA_SIZE, -1)
	if rect.Empty() {
		return gocv.NewMat(), fmt.Errorf("hlavny obdlznik nebol najdeny")
	}

	// Vyrezeme pas nad hlavnym obdlznikom.
	// Sirka ostane rovnaka ako hlavny obdlznik, vyska je len male pasmo nad nim.
	headerRect := image.Rectangle{
		Min: image.Point{
			X: rect.Min.X + SUBGROUP_SIDE_PADDING,
			Y: rect.Min.Y - SUBGROUP_HEADER_HEIGHT,
		},
		Max: image.Point{
			X: rect.Max.X - SUBGROUP_SIDE_PADDING,
			Y: rect.Min.Y - SUBGROUP_BOTTOM_PADDING,
		},
	}

	// Ochrana proti vybehnutiu mimo obrazka
	if headerRect.Min.X < 0 {
		headerRect.Min.X = 0
	}
	if headerRect.Min.Y < 0 {
		headerRect.Min.Y = 0
	}
	if headerRect.Max.X > mat.Cols() {
		headerRect.Max.X = mat.Cols()
	}
	if headerRect.Max.Y > mat.Rows() {
		headerRect.Max.Y = mat.Rows()
	}

	if headerRect.Dx() <= 0 || headerRect.Dy() <= 0 {
		return gocv.NewMat(), fmt.Errorf("oblast headera je prilis mala alebo mimo obrazka")
	}

	croppedMat := mat.Region(headerRect)
	SaveMat("./assets/group-crop2.png", croppedMat)
	return croppedMat, nil
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
func GetQuestionNumber(mat *gocv.Mat, i int, numberOfChoices int) int {
	errorLogger := logging.GetErrorLogger()
	rect := image.Rectangle{Min: image.Point{PADDING, PADDING + (i * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE)}, Max: image.Point{(mat.Cols() / (numberOfChoices + 1)) - PADDING, ((i + 1) * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE) - PADDING}}
	questionMat := mat.Region(rect)
	path := fmt.Sprintf("./assets/nemberMat_%d.png", i)
	SaveMat(path, questionMat)
	defer questionMat.Close()
	SaveMat(TEMP_IMAGE_PATH, questionMat)
	questionNum, err := ocr.ExtractQuestionNumber(TEMP_IMAGE_PATH)
	files.DeleteFile(TEMP_IMAGE_PATH)

	if err != nil {
		errorLogger.Error("Chyba pri extrakcii čísla otázky",
			slog.Int("questionIndex", i),
			slog.String("error", err.Error()),
			slog.Int("questionNum", questionNum),
		)
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
func GetAnswer(mat *gocv.Mat, i int, numberOfChoices int) rune {
	answer := rune('x')
	state := StateEmpty
	pad := checkboxPaddingForChoices(numberOfChoices)
	for j := 1; j <= numberOfChoices; j++ {
		padding := CHECKBOX_AREA_PADDING
		if i == 0 || i == NUMBER_OF_QUESTIONS_PER_PAGE-1 {
			padding = 0
		}
		checkbox := image.Rectangle{Min: image.Point{(mat.Cols() / (numberOfChoices + 1) * (j)), padding + (i * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE)}, Max: image.Point{(mat.Cols() / (numberOfChoices + 1)) * (j + 1), ((i + 1) * mat.Rows() / NUMBER_OF_QUESTIONS_PER_PAGE) - padding}}
		checkboxMat := mat.Region(checkbox)
		//rect := FindRectangle(&checkboxMat, ANSWER_SQUARE_MIN_AREA_SIZE, ANSWER_SQUARE_MAX_AREA_SIZE)
		minArea, maxArea := answerSquareAreaBounds(numberOfChoices)
		rect := FindRectangle(&checkboxMat, minArea, maxArea)
		if rect.Empty() {
			checkboxMat.Close()
			if state == StateCircleFound {
				return rune('x')
			}
			answer = rune('a' + (j - 1))
			state = StateCircleFound
			continue
		}
		checkboxWithoutBorder := image.Rectangle{Min: image.Point{rect.Min.X + pad, rect.Min.Y + pad}, Max: image.Point{rect.Max.X - pad, rect.Max.Y - pad}}
		if checkboxWithoutBorder.Dx() <= 0 || checkboxWithoutBorder.Dy() <= 0 {
			checkboxMat.Close()
			continue
		}
		rectMat := checkboxMat.Region(checkboxWithoutBorder)
		meanIntensity := rectMat.Mean()
		if meanIntensity.Val1 < MEAN_INTENSITY_X_HIGHEST && meanIntensity.Val1 > MEAN_INTENSITY_X_LOWEST {
			rectMat.Close()
			checkboxMat.Close()
			if state == StateEmpty {
				answer = rune('a' + (j - 1))
				state = StateXFound
				continue
			} else if state == StateXFound {
				answer = rune('x')
			}
		}
		//fmt.Println(meanIntensity.Val1)
		rectMat.Close()
		checkboxMat.Close()
	}
	return answer
}

func GetGroupCode(headerMat *gocv.Mat) rune {
	state := StateEmpty
	answer := rune('x')
	numberOfGroups := 8
	pad := checkboxPaddingForChoices(numberOfGroups)
	for j := 0; j < numberOfGroups; j++ {
		checkbox := image.Rectangle{
			Min: image.Point{
				X: j * headerMat.Cols() / numberOfGroups,
				Y: 0,
			},
			Max: image.Point{
				X: (j + 1) * headerMat.Cols() / numberOfGroups,
				Y: headerMat.Rows(),
			},
		}

		checkboxMat := headerMat.Region(checkbox)
		minArea, maxArea := answerSquareAreaBounds(numberOfGroups)
		rect := FindRectangle(&checkboxMat, minArea, maxArea)
		if rect.Empty() {
			checkboxMat.Close()
			if state == StateCircleFound {
				return rune('x')
			}
			answer = rune('a' + j)
			state = StateCircleFound
			continue
		}
		checkboxWithoutBorder := image.Rectangle{Min: image.Point{rect.Min.X + pad, rect.Min.Y + pad}, Max: image.Point{rect.Max.X - pad, rect.Max.Y - pad}}
		if checkboxWithoutBorder.Dx() <= 0 || checkboxWithoutBorder.Dy() <= 0 {
			checkboxMat.Close()
			continue
		}
		rectMat := checkboxMat.Region(checkboxWithoutBorder)
		meanIntensity := rectMat.Mean()
		if meanIntensity.Val1 < MEAN_INTENSITY_X_HIGHEST && meanIntensity.Val1 > MEAN_INTENSITY_X_LOWEST {
			rectMat.Close()
			checkboxMat.Close()
			if state == StateEmpty {
				answer = rune('a' + j)
				state = StateXFound
				continue
			} else if state == StateXFound {
				answer = rune('x')
			}
		}

		/*path := fmt.Sprintf("./assets/group_part_%d.png", i)
		SaveMat(path, checkboxMat)*/

		rectMat.Close()
		checkboxMat.Close()
	}
	return answer
}

func GetSubGroupCode(headerMat *gocv.Mat) int {
	state := StateEmpty
	answer := -1
	numberOfSubgroups := 5
	pad := CHECKBOX_PADDING

	for j := 0; j < numberOfSubgroups; j++ {
		checkbox := image.Rectangle{
			Min: image.Point{
				X: j * headerMat.Cols() / numberOfSubgroups,
				Y: 0,
			},
			Max: image.Point{
				X: (j + 1) * headerMat.Cols() / numberOfSubgroups,
				Y: headerMat.Rows(),
			},
		}

		checkboxMat := headerMat.Region(checkbox)

		minArea, maxArea := answerSquareAreaBounds(numberOfSubgroups)
		rect := FindRectangle(&checkboxMat, minArea, maxArea)
		if rect.Empty() {
			checkboxMat.Close()
			if state == StateCircleFound {
				return -1
			}
			answer = j + 1
			state = StateCircleFound
			continue
		}

		checkboxWithoutBorder := image.Rectangle{
			Min: image.Point{rect.Min.X + pad, rect.Min.Y + pad},
			Max: image.Point{rect.Max.X - pad, rect.Max.Y - pad},
		}

		if checkboxWithoutBorder.Dx() <= 0 || checkboxWithoutBorder.Dy() <= 0 {
			checkboxMat.Close()
			continue
		}

		rectMat := checkboxMat.Region(checkboxWithoutBorder)
		meanIntensity := rectMat.Mean()

		if meanIntensity.Val1 < MEAN_INTENSITY_X_HIGHEST && meanIntensity.Val1 > MEAN_INTENSITY_X_LOWEST {
			rectMat.Close()
			checkboxMat.Close()

			if state == StateEmpty {
				answer = j + 1
				state = StateXFound
				continue
			} else if state == StateXFound {
				return -1
			}
		}

		rectMat.Close()
		checkboxMat.Close()
	}

	return answer
}
