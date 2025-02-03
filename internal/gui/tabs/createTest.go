package tabs
/*
TODO: 
dorobit ulozenie do databazy
dorobit scroll pri generovani otazok 
povolit iba jednu moznost pri danej otazke
osetrit vstupy
*/
import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/unit"
	"fmt"
	"strconv"
	"ScanEvalApp/internal/logging"
	"log/slog"
)

var (
	nameInput      widget.Editor
	roomInput      widget.Editor
	timeInput      widget.Editor
	questionsInput widget.Editor
	submitButton   widget.Clickable
	createButton   widget.Clickable
	questionForms  []questionForm
	showQuestions  bool
)

var questionList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

type questionForm struct {
	selectedOption widget.Enum // Uchováva vybranú možnosť (A, B, C, D, E)
}

// CreateTest renders the content for the "Vytvorenie Písomky" tab.
func CreateTest(gtx layout.Context, th *material.Theme) layout.Dimensions {
	logger := logging.GetLogger()

	if createButton.Clicked(gtx) {
		logger.Info("Kliknutie na tlačidlo Vytvoriť test")
		if questionsInput.Text() != "" {
			n := parseNumber(questionsInput.Text())
			if n > 0 {
				updateQuestionForms(n)
				showQuestions = true
			}
		}
	}

	return layout.Flex{
		Axis:    layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 4, Bottom: 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Editor(th, &nameInput, "Názov                  ").Layout(gtx)
						})
					}),
					
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Editor(th, &roomInput, "Miestnosť").Layout(gtx)
						})
					}),
					
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Editor(th, &timeInput, "Čas   ").Layout(gtx)
						})
					}),
					
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Editor(th, &questionsInput, "Počet otázok").Layout(gtx)
						})
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Button(th, &createButton, "Vytvoriť test").Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if showQuestions {
				btn := material.Button(th, &submitButton, "Odoslať")
				if submitButton.Clicked(gtx) {
					logger.Info("Kliknutie na tlačidlo Odoslať")
					submitForm()
				}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if showQuestions {
				return material.List(th, &questionList).Layout(gtx, len(questionForms), func(gtx layout.Context, i int) layout.Dimensions {
					qf := &questionForms[i]
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceAround,
					}.Layout(gtx, renderOptions(gtx, th, i+1, qf)...)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

func parseNumber(input string) int {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	num, err := strconv.Atoi(input)
	if err != nil {
		errorLogger.Error("Chyba pri parsovaní počtu otázok", slog.String("error", err.Error()))
	} else {
		logger.Debug("Preparsované číslo pocet otazok:", slog.Int("count", num))
	}
	return num
}

func updateQuestionForms(n int) {
	for len(questionForms) < n {
		questionForms = append(questionForms, questionForm{})
	}
	for len(questionForms) > n {
		questionForms = questionForms[:len(questionForms)-1]
	}
}

func renderQuestionForms(gtx layout.Context, th *material.Theme) []layout.FlexChild {
	children := make([]layout.FlexChild, len(questionForms))
	for i := range questionForms { // Prechádzame len indexy, aby sme pracovali priamo so slice-om
		qf := &questionForms[i] // Uložíme si pointer na konkrétny prvok
		children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceAround,
			}.Layout(gtx, renderOptions(gtx, th, i+1, qf)...) // Odovzdávame pointer na správny prvok
		})
	}
	return children
}



func renderOptions(gtx layout.Context, th *material.Theme, questionIndex int, qf *questionForm) []layout.FlexChild {
	options := []string{"A", "B", "C", "D", "E"}
	children := make([]layout.FlexChild, len(options)+1) // Prvý prvok je číslo otázky

	// Pridáme číslo otázky (napr. "01:")
	children[0] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return material.Label(th, unit.Sp(15), fmt.Sprintf("%02d:", questionIndex)).Layout(gtx)
	})

	// Vykreslíme rádio tlačidlá pre možnosti A–E
	for i, option := range options {
		i, option := i, option
		children[i+1] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.RadioButton(th, &qf.selectedOption, option, option).Layout(gtx)
		})
	}

	return children
}

func submitForm() {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Načítame údaje zo všetkých inputov
	nazov := nameInput.Text()
	miestnost := roomInput.Text()
	cas := timeInput.Text()
	pocetOtazok, err := strconv.Atoi(questionsInput.Text())
	if err != nil {
		errorLogger.Error("Chyba pri parsovaní počtu otázok", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return
	}

	// Vytvoríme si výsledný výpis
	logger.Info("Formulár odoslaný", 
		slog.String("nazov", nazov), 
		slog.String("miestnost", miestnost), 
		slog.String("cas", cas), 
		slog.Int("pocetOtazok", pocetOtazok))
		// Premenná pre uchovávanie zaškrtnutých možností
//	var selectedOptions string
	// Prejdeme každú otázku a jej odpovede
/*	for i, qf := range questionForms {
		fmt.Printf("otazka c. %d", i)
		// Pre každú možnosť otázky (A, B, C, D, E) skontrolujeme, či je zaškrtnutá
		for j, option := range qf.options {
			optionStr := string(rune('A' + j))
			if option.Value {
				// Ak je možnosť zaškrtnutá, pridáme ju do stringu
				if selectedOptions != "" {
					selectedOptions += ", "
				}
				selectedOptions += optionStr
				fmt.Printf("  - Zaškrtnuté: %s\n", optionStr)
			}
		}

		// Ak sú zaškrtnuté možnosti, vypíšeme ich

	}
	if selectedOptions != "" {
		fmt.Printf("  - Zaškrtnuté možnosti: %s\n", selectedOptions)
	} else {
		fmt.Println("  - Žiadna možnosť nie je zaškrtnutá")
	}
	

*/
}
