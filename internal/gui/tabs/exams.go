package tabs

import (
	"ScanEvalApp/internal/database/repository"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"fmt"
	"gorm.io/gorm"
	"gioui.org/widget"
	//"gioui.org/unit"
	"ScanEvalApp/internal/logging"
	"log/slog"
    "ScanEvalApp/internal/database/models"
    "ScanEvalApp/internal/gui/tabmanager" 
)
var deleteButtons []widget.Clickable
var showAnsButtons []widget.Clickable
var evaluateTestBtns []widget.Clickable
// scrollovanie
var examList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

// Exams renders the "Exams" tab with dynamically generated columns based on data from the database.
func Exams(gtx layout.Context, th *material.Theme, selectedTestID *uint, db *gorm.DB, tm *tabmanager.TabManager) layout.Dimensions {
    //logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

    tests, err := repository.GetAllTests(db)
    if err != nil {
        errorLogger.Error("Chyba pri načítaní testov", slog.String("error", err.Error()))
        return layout.Dimensions{}
    }

    columns := []string{"Názov", "Rok", "Počet otázok", "Počet študentov", "Dátum", "Ukázať odpovede", "Vymazať", "Vyhodnotiť"}
    columnWidths := []float32{0.2, 0.15, 0.15, 0.1, 0.1, 0.1, 0.1, 0.1} // Pomery šírok
    if len(deleteButtons) != len(tests) {
		deleteButtons = make([]widget.Clickable, len(tests))
	}
    if len(showAnsButtons) != len(tests) {
		showAnsButtons = make([]widget.Clickable, len(tests))
	}
    if len(evaluateTestBtns) != len(tests) {
		evaluateTestBtns = make([]widget.Clickable, len(tests))
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
                layout.Flexed(columnWidths[7], func(gtx layout.Context) layout.Dimensions {
                    return material.Body1(th, columns[7]).Layout(gtx)
                }),
                
            )
        }),
        layout.Rigid(func(gtx layout.Context) layout.Dimensions {
            return material.List(th, &examList).Layout(gtx, len(tests), func(gtx layout.Context, i int) layout.Dimensions {
                test := tests[i]
                if deleteButtons[i].Clicked(gtx) {
                    deleteTest(test.ID, db)
                    tests = removeTestFromList(tests, i) // Remove test from the list for UI update
                }
                if showAnsButtons[i].Clicked(gtx) {
                    showAnsTest(test.ID)
                    *selectedTestID = test.ID  // Nastavenie ID testu
                    tm.ActiveTab = 3          // Prechod na UploadTab
                
                }
                if evaluateTestBtns[i].Clicked(gtx) {
                    *selectedTestID = test.ID  // Nastavenie ID testu
                    tm.ActiveTab = 3          // Prechod na UploadTab
                
                }
        
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
                        return material.Body1(th, fmt.Sprintf("datum")).Layout(gtx)
                    }),
                    layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
                        btn := material.Button(th, &showAnsButtons[i], "Zobraziť")
                        return btn.Layout(gtx)
                    }),
                    layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
                        btn := material.Button(th, &deleteButtons[i], "Vymazať")
                        return btn.Layout(gtx)
                    }),
                    layout.Flexed(columnWidths[7], func(gtx layout.Context) layout.Dimensions {
                        btn := material.Button(th, &evaluateTestBtns[i], "Vyhodnotiť")
                        return btn.Layout(gtx)
                    }),
                )
            })
        }),        
    )
}


func deleteTest(Id uint, db *gorm.DB) {
    logger := logging.GetLogger()
    errorLogger := logging.GetErrorLogger()

    // Deleting the test from the database
    if err := repository.DeleteTest(db, Id); err != nil {
        errorLogger.Error("Chyba pri vymazávaní testu", slog.Uint64("ID", uint64(Id)), slog.String("error", err.Error()))
        return
    }

    logger.Info("Vymazanie testu s ID", slog.Uint64("ID", uint64(Id)))
}

func removeTestFromList(tests []models.Test, index int) []models.Test {
    // Removing the test from the list at the specified index
    return append(tests[:index], tests[index+1:]...)
}

func showAnsTest(Id uint) {
    logger := logging.GetLogger()

    logger.Info("Ukázanie opovedí testu s ID", slog.Uint64("ID", uint64(Id)))
}
