package scanprocessing

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// DetectAnswersOnSheet spracuje už narovnaný a zosivatený harok (mat),
// vyreže blok s odpoveďami, prejde riadky a vráti odpovede.
// choices = počet možností (napr. 5), questions = počet otázok v hárku (napr. 20).
func DetectAnswersOnSheet(mat *gocv.Mat, choices, questions int) ([]rune, error) {
	if choices < 1 || questions < 1 {
		return nil, fmt.Errorf("choices a questions musia byť > 0")
	}
	cropped := CropMatAnswersOnly(mat) // použitá existujúca funkcia (vracia nový Mat výrezu)
	results := make([]rune, 0, questions)

	for i := 0; i < questions; i++ {
		ans := getAnswerParametric(&cropped, i, choices, questions)
		results = append(results, ans)
	}

	// nahradíme pôvodný mat vyrezaným (rovnako ako EvaluateAnswers)
	*mat = cropped
	return results, nil
}

// getAnswerParametric je parametrická verzia GetAnswer,
// používa choices a questions namiesto konštánt.
func getAnswerParametric(mat *gocv.Mat, rowIndex, choices, questions int) rune {
	choices = 8
	answer := rune('x')
	state := StateEmpty

	for j := 1; j <= choices; j++ {
		padding := CHECKBOX_AREA_PADDING
		if rowIndex == 0 || rowIndex == questions-1 {
			padding = 0
		}
		stepPx := 138
		xMin := j * stepPx
		xMax := (j + 1) * stepPx
		if xMin < 0 {
			xMin = 0
		}
		if xMax > mat.Cols() {
			xMax = mat.Cols()
		}
		checkbox := image.Rectangle{
			Min: image.Point{xMin, padding + (rowIndex * mat.Rows() / questions)},
			Max: image.Point{xMax, ((rowIndex + 1) * mat.Rows() / questions) - padding},
		}
		checkboxMat := mat.Region(checkbox)
		rect := FindRectangle(&checkboxMat, ANSWER_SQUARE_MIN_AREA_SIZE, ANSWER_SQUARE_MAX_AREA_SIZE)

		if rect.Empty() {
			// žiadny rámik → interpretuj ako krúžok/označenie v stĺpci bez rámika
			if state == StateCircleFound {
				checkboxMat.Close()
				return rune('x') // druhé označenie -> neplatné
			}
			answer = rune('a' + (j - 1))
			state = StateCircleFound
			checkboxMat.Close()
			continue
		}

		inner := image.Rectangle{
			Min: image.Point{rect.Min.X + CHECKBOX_PADDING, rect.Min.Y + CHECKBOX_PADDING},
			Max: image.Point{rect.Max.X - CHECKBOX_PADDING, rect.Max.Y - CHECKBOX_PADDING},
		}
		rectMat := checkboxMat.Region(inner)
		mean := rectMat.Mean()

		if mean.Val1 < MEAN_INTENSITY_X_HIGHEST && mean.Val1 > MEAN_INTENSITY_X_LOWEST {
			if state == StateEmpty {
				answer = rune('a' + (j - 1))
				state = StateXFound
			} else if state == StateXFound {
				answer = rune('x') // druhé X v riadku -> neplatné
			}
		}

		rectMat.Close()
		checkboxMat.Close()
	}
	return answer
}

// Pomocná funkcia na vypísanie
func PrintAnswers(answers []rune) {
	fmt.Println(string(answers))
}

// Voliteľný wrapper: spraví aj konverziu/zarovnanie pre raw image.
// img je RGBA stránka (napr. z PDF), zavolá sa konverzia a spracovanie.
func DetectAnswersFromImage(img *image.RGBA, choices, questions int) ([]rune, error) {
	mat := ImageToMat(img)
	defer mat.Close()
	mat = MatToGrayscale(mat)
	mat = FixImageRotation(mat)
	return DetectAnswersOnSheet(&mat, choices, questions)
}
