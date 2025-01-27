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

type questionForm struct {
	options []widget.Bool
}

// CreateTest renders the content for the "Vytvorenie Písomky" tab.
func CreateTest(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if createButton.Clicked(gtx) {
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
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx, renderQuestionForms(gtx, th)...) // Render questions.
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if showQuestions {
				btn := material.Button(th, &submitButton, "Odoslať")
				if submitButton.Clicked(gtx) {
					submitForm()
				}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
	)
}

func parseNumber(input string) int {
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Chyba pri parsovaní:", err)
	} else {
		fmt.Println("Preparsované číslo pocet otazok:", num)
	}
	return num
}

func updateQuestionForms(n int) {
	for len(questionForms) < n {
		questionForms = append(questionForms, questionForm{
			options: make([]widget.Bool, 5), // A, B, C, D, E
		})
	}
	for len(questionForms) > n {
		questionForms = questionForms[:len(questionForms)-1]
	}
}
func renderQuestionForms(gtx layout.Context, th *material.Theme) []layout.FlexChild {
	children := make([]layout.FlexChild, len(questionForms))
	for i, qf := range questionForms {
		i, qf := i, qf
		children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceAround,
			}.Layout(gtx, renderOptions(gtx, th, i+1, qf.options)...) // Pass question number (i+1) to renderOptions
		})
		
	}
	return children
}

func renderOptions(gtx layout.Context, th *material.Theme, questionIndex int, options []widget.Bool) []layout.FlexChild {
	children := make([]layout.FlexChild, len(options)+1) // Add space for the question number
	// Add the question number label
	children[0] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return material.Label(th, unit.Sp(15), fmt.Sprintf("%02d:", questionIndex)).Layout(gtx)
	})
	
	// Render the options (A, B, C, D, E)
	for i := range options {
		i := i
		children[i+1] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(th, &options[i], string(rune('A'+i))).Layout(gtx)
		})
	}
	return children
}

func submitForm() {
	// Načítame údaje zo všetkých inputov
	nazov := nameInput.Text()
	miestnost := roomInput.Text()
	cas := timeInput.Text()
	pocetOtazok, err := strconv.Atoi(questionsInput.Text())
	if err != nil {
		fmt.Println("Chyba pri parsovaní počtu otázok:", err)
		return
	}

	// Vytvoríme si výsledný výpis
	fmt.Println("Názov:", nazov)
	fmt.Println("Miestnosť:", miestnost)
	fmt.Println("Čas:", cas)
	fmt.Println("Počet otázok:", pocetOtazok)
		// Premenná pre uchovávanie zaškrtnutých možností
	var selectedOptions string
	// Prejdeme každú otázku a jej odpovede
	for i, qf := range questionForms {
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
	


}
