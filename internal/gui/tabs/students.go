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

	"gioui.org/unit"
)

// Tlačidlo na tlač všetkých hárkov

var printButtons []widget.Clickable
var searchQuery widget.Editor

// scrollovanie
var studentList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

// StudentsTab renders the "Students" tab with a table of students.
func Students(gtx layout.Context, th *material.Theme, db *gorm.DB) layout.Dimensions {
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
			margin := unit.Dp(16) // Nastav si veľkosť marginu podľa potreby
			gtx.Constraints.Min.X = 0
			gtx.Constraints.Min.Y = 0
			gtx.Constraints.Max.X -= 2 * gtx.Dp(margin)

			inset := layout.Inset{
				Top:    margin,
				Bottom: margin,
				Left:   margin,
			}

			editor := material.Editor(th, &searchQuery, "Vyhľadávanie (Meno, Priezvisko, Registračné číslo)")
			return inset.Layout(gtx, editor.Layout)
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[0]).Layout(gtx)
				}),
				layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[1]).Layout(gtx)
				}),
				layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[2]).Layout(gtx)
				}),
				layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[3]).Layout(gtx)
				}),
				layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[4]).Layout(gtx)
				}),
				layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[5]).Layout(gtx)
				}),
				layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
					return material.Body1(th, columns[6]).Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.List(th, &studentList).Layout(gtx, len(students), func(gtx layout.Context, i int) layout.Dimensions {
				student := students[i]
				if printButtons[i].Clicked(gtx) {
					printSheet(student.RegistrationNumber)
				}

				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, student.Name).Layout(gtx)
					}),
					layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, student.Surname).Layout(gtx)
					}),
					layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, student.BirthDate.Format("2006-01-02")).Layout(gtx)
					}),
					layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, student.RegistrationNumber).Layout(gtx)
					}),
					layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, student.Room).Layout(gtx)
					}),
					layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
						return material.Body1(th, fmt.Sprintf("%d", student.Score)).Layout(gtx)
					}),
					layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &printButtons[i], "Tlačiť hárok")
						return btn.Layout(gtx)
					}),
				)
			})
		}),
	)
}

func printSheet(registrationNumber string) {
	logger := logging.GetLogger()

	logger.Info("Volám tlač hárku pre študenta ID", slog.String("ID", registrationNumber))
}
