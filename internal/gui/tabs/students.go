package tabs

import (
	"ScanEvalApp/internal/database/repository"
	"fmt"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gorm.io/gorm"
)
	// Tlačidlo na tlač všetkých hárkov
var printAllButton widget.Clickable
// StudentsTab renders the "Students" tab with a table of students.
func Students(gtx layout.Context, th *material.Theme, db *gorm.DB) layout.Dimensions {
	students, err := repository.GetAllStudents(db)
	if err != nil {
		fmt.Println("Chyba pri načítaní študentov:", err)
		return layout.Dimensions{}
	}

	columns := []string{"Meno", "Priezvisko", "Dátum narodenia", "Registračné číslo", "Miestnosť", "Skóre", "Tlačiť hárok"}
	columnWidths := []float32{0.15, 0.15, 0.2, 0.2, 0.1, 0.1, 0.1} // Pomery šírok



	

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(th, &printAllButton, "Tlačiť všetky hárky")
			if printAllButton.Clicked(gtx) {
				printAllSheets()
			}
			return btn.Layout(gtx)
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
			var rows []layout.FlexChild
			for _, student := range students {
				/*printButton := new(widget.Clickable)

				if printButton.Clicked(gtx) {
					printSheet(student.ID)
				}*/
				rows = append(rows,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
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
								return material.Body1(th, fmt.Sprintf("Tlacit harok")).Layout(gtx)
							}),
						)
					}),
				)
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
		}),
	)
}


func printAllSheets() {
	fmt.Println("Volám tlač všetky hárky")
}

func printSheet(studentID uint) {
	fmt.Printf("Volám tlač hárku pre študenta ID: %d\n", studentID)
}