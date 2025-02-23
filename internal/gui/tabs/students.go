package tabs

import (
	"ScanEvalApp/internal/database/repository"
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gorm.io/gorm"

	//"reflect"
	"ScanEvalApp/internal/logging"
	"log/slog"
	"ScanEvalApp/internal/gui/widgets"
	"gioui.org/unit"
	//"ScanEvalApp/internal/gui/themeUI"
	themeIU "ScanEvalApp/internal/gui/themeUI"
)

// Tlačidlo na tlač všetkých hárkov

var printButtons []widget.Clickable
var searchQuery widget.Editor

// scrollovanie
var studentList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

// StudentsTab renders the "Students" tab with a table of students.
func Students(gtx layout.Context, th *themeIU.Theme, db *gorm.DB) layout.Dimensions {
	//logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	students, err := repository.GetAllStudents(db)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní študentov", slog.String("error", err.Error()))
		return layout.Dimensions{}
	}
	// Filtrovanie študentov na základe textu v searchQuery
	query := searchQuery.Text()

	// Ak je query nenulové, filtrujeme podľa mena, priezviska a registračného čísla
	if query != "" {
		students, err = repository.GetStudentsQuery(db, query)
		if err != nil {
			errorLogger.Error("Chyba pri načítaní študentov", slog.String("error", err.Error()))
			return layout.Dimensions{}
		}
	} else {
		students, err = repository.GetAllStudents(db)
		if err != nil {
			errorLogger.Error("Chyba pri načítaní študentov", slog.String("error", err.Error()))
			return layout.Dimensions{}
		}
	}
	columns := []string{"Meno", "Priezvisko", "Dátum narodenia", "Registračné číslo", "Miestnosť", "Skóre", "Tlačiť hárok"}
	columnWidths := []float32{0.15, 0.15, 0.2, 0.2, 0.1, 0.1, 0.1} // Pomery šírok
	if len(printButtons) != len(students) {
		printButtons = make([]widget.Clickable, len(students))
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := widgets.NewEditorField(th.Theme, &searchQuery, "Vyhľadávanie (Meno, Priezvisko, Registračné číslo)") // Šírku riadi columnWidths
				return editor.Layout(gtx, th)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[0]).Layout(gtx)
					}),
					layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[1]).Layout(gtx)
					}),
					layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[2]).Layout(gtx)
					}),
					layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[3]).Layout(gtx)
					}),
					layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[4]).Layout(gtx)
					}),
					layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[5]).Layout(gtx)
					}),
					layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th.Material(), columns[6]).Layout(gtx)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(15), Right: unit.Dp(15)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.List(th.Material(), &studentList).Layout(gtx, len(students), func(gtx layout.Context, i int) layout.Dimensions {
					student := students[i]
					if printButtons[i].Clicked(gtx) {
						printSheet(student.RegistrationNumber)
					}

					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
							return material.Body1(th.Material(), student.Name).Layout(gtx)
						}),
						layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
							return material.Body1(th.Material(), student.Surname).Layout(gtx)
						}),
						layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
							return material.Body1(th.Material(), student.BirthDate.Format("2006-01-02")).Layout(gtx)
						}),
						layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
							return material.Body1(th.Material(), student.RegistrationNumber).Layout(gtx)
						}),
						layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
							return material.Body1(th.Material(), student.Room).Layout(gtx)
						}),
						layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
							return material.Body1(th.Material(), fmt.Sprintf("%d", student.Score)).Layout(gtx)
						}),
						layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(th.Material(), &printButtons[i], "Tlačiť hárok")
							return btn.Layout(gtx)
						}),
					)
				})
			})
		}),
	)
}

func printSheet(registrationNumber string) {
	logger := logging.GetLogger()

	logger.Info("Volám tlač hárku pre študenta ID", slog.String("ID", registrationNumber))
}
