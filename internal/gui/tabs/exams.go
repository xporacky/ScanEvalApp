package tabs

import (
	"ScanEvalApp/internal/database/repository"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"fmt"
	"gorm.io/gorm"
	//"gioui.org/widget"
	"gioui.org/unit"

)

// Exams renders the "Exams" tab with dynamically generated columns based on data from the database.
func Exams(gtx layout.Context, th *material.Theme, db *gorm.DB) layout.Dimensions {
	// Načítame všetky testy z databázy
	tests, err := repository.GetAllTests(db)
	if err != nil {
		fmt.Println("Chyba pri načítaní testov:", err)
		return layout.Dimensions{}
	}

	// Predpokladajme, že máme pre každý test tieto vlastnosti (dáta z DB môžu byť rôzne)
	columns := []string{"Názov", "Rok", "Počet otázok", "Počet študentov", "Miestnosť", "Ukázať odpovede", "Vymazať"}

	// Vytvoríme flex layout na vykreslenie tabulky
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Vykreslenie prvého riadku tabuľky
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Názov testu (prvý stĺpec)
					label := material.Body1(th, columns[0])
					return label.Layout(gtx)
				}),
				// Pridáme "medzeru" (simulujeme čiaru)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Rok testu (druhý stĺpec)
					label := material.Body1(th, columns[1])
					return label.Layout(gtx)
				}),
				// "Medzera" medzi stĺpcami
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Počet otázok (tretí stĺpec)
					label := material.Body1(th, columns[2])
					return label.Layout(gtx)
				}),
				// "Medzera" medzi stĺpcami
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Počet študentov (štvrtý stĺpec)
					label := material.Body1(th, columns[3])
					return label.Layout(gtx)
				}),
				// "Medzera" medzi stĺpcami
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Miestnosť (piaty stĺpec)
					label := material.Body1(th, columns[4])
					return label.Layout(gtx)
				}),
				// "Medzera" medzi stĺpcami
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Miestnosť (siesty stĺpec)
					label := material.Body1(th, columns[5])
					return label.Layout(gtx)
				}),
				// "Medzera" medzi stĺpcami
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Miestnosť (siedmi stĺpec)
					label := material.Body1(th, columns[6])
					return label.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var rows []layout.FlexChild
			for _, test := range tests {
				rows = append(rows,
					// Vykreslíme celý riadok
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							// Názov testu (prvý stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, test.Title)
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
							}),
							// Rok testu (druhý stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, test.SchoolYear)
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
							}),
							// Počet otázok (tretí stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, fmt.Sprintf("%d", test.QuestionCount))
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
							}),
							// Počet študentov (štvrtý stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, fmt.Sprintf("%d", len(test.Students)))
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
							}),
							// Miestnosť (piaty stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, test.Room)
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
							}),
							// Ukázať odpovede (šiesti stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, "Ukázať")
								return label.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Spacer{Width: unit.Dp(10)}.Layout(gtx)
							}),
							// Vymazať (siedmy stĺpec)
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := material.Body1(th, "Vymazať")
								return label.Layout(gtx)
							}),
						)
					}),
				)
			}
			// Layout všetkých riadkov
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
		}),
	)
}