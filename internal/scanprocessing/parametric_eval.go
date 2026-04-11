package scanprocessing

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Adjusts the padding for checkboxes based on the number of choices.
func checkboxPaddingForChoices(choices int) int {
	if choices == 8 {
		return 4
	}
	return CHECKBOX_PADDING
}

// DetectAnswersOnSheet spracuje už narovnaný a zosivatený harok (mat),
// vyreže blok s odpoveďami, prejde riadky a vráti odpovede.
// choices = počet možností (napr. 5), questions = počet otázok v hárku (napr. 20).
func DetectAnswersOnSheet(mat *gocv.Mat, choices, questions int) ([]rune, error) {
	if choices < 1 || questions < 1 {
		return nil, fmt.Errorf("choices a questions musia byť > 0")
	}
	cropped, err := CropMatAnswersOnly(mat)
	if err != nil {
		return nil, fmt.Errorf("prazdna alebo neplatna strana: %w", err)
	}
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
	answer := rune('x')
	state := StateEmpty

	for j := 1; j <= choices; j++ {
		padding := CHECKBOX_AREA_PADDING
		if rowIndex == 0 || rowIndex == questions-1 {
			padding = 0
		}
		checkbox := image.Rectangle{
			Min: image.Point{(mat.Cols() / (choices + 1) * (j)), padding + (rowIndex * mat.Rows() / questions)},
			Max: image.Point{(mat.Cols() / (choices + 1)) * (j + 1), ((rowIndex + 1) * mat.Rows() / questions) - padding},
		}
		checkboxMat := mat.Region(checkbox)
		minArea, maxArea := answerSquareAreaBounds(choices)
		rect := FindRectangle(&checkboxMat, minArea, maxArea)

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
		pad := checkboxPaddingForChoices(choices)
		inner := image.Rectangle{
			Min: image.Point{rect.Min.X + pad, rect.Min.Y + pad},
			Max: image.Point{rect.Max.X - pad, rect.Max.Y - pad},
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
