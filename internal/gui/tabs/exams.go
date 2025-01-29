package tabs

import (
	"ScanEvalApp/internal/database/repository"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"fmt"
	"gorm.io/gorm"
	"gioui.org/widget"
	//"gioui.org/unit"

)
var deleteButtons []widget.Clickable
var showAnsButtons []widget.Clickable
// Exams renders the "Exams" tab with dynamically generated columns based on data from the database.
func Exams(gtx layout.Context, th *material.Theme, db *gorm.DB) layout.Dimensions {
    tests, err := repository.GetAllTests(db)
    if err != nil {
        fmt.Println("Chyba pri načítaní testov:", err)
        return layout.Dimensions{}
    }

    columns := []string{"Názov", "Rok", "Počet otázok", "Počet študentov", "Miestnosť", "Ukázať odpovede", "Vymazať"}
    columnWidths := []float32{0.2, 0.15, 0.15, 0.15, 0.15, 0.1, 0.1} // Pomery šírok
    if len(deleteButtons) != len(tests) {
		deleteButtons = make([]widget.Clickable, len(tests))
	}
    if len(showAnsButtons) != len(tests) {
		showAnsButtons = make([]widget.Clickable, len(tests))
	}

    return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
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
            for idx, test := range tests {
                if deleteButtons[idx].Clicked(gtx) {
					deleteTest(test.ID)
				}
                if showAnsButtons[idx].Clicked(gtx) {
					showAnsTest(test.ID)
				}
                rows = append(rows,
                    layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                        return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
                            layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
                                return material.Body1(th, test.Title).Layout(gtx)
                            }),
                            layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
                                return material.Body1(th, test.SchoolYear).Layout(gtx)
                            }),
                            layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
                                return material.Body1(th, fmt.Sprintf("%d", test.QuestionCount)).Layout(gtx)
                            }),
                            layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
                                return material.Body1(th, fmt.Sprintf("%d", len(test.Students))).Layout(gtx)
                            }),
                            layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
                                return material.Body1(th, test.Room).Layout(gtx)
                            }),
                            layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
								// Button to print student sheet
								btn := material.Button(th, &showAnsButtons[idx], "Zobraziť")
								return btn.Layout(gtx)
							}),
                            layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
								// Button to print student sheet
								btn := material.Button(th, &deleteButtons[idx], "Vymazať")
								return btn.Layout(gtx)
							}),
                        )
                    }),
                )
            }
            return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
        }),
    )
}


func deleteTest(Id uint) {
	fmt.Printf("delete testu s ID: %d\n", Id)
}

func showAnsTest(Id uint) {
	fmt.Printf("show answer testu s ID: %d\n", Id)
}
