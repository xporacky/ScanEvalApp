package tabs

import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/gui/tabmanager"
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"
	"ScanEvalApp/internal/logging"
	"fmt"
	"log/slog"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gorm.io/gorm"
)

var deleteButtons []widget.Clickable
var showAnsButtons []widget.Clickable
var evaluateTestBtns []widget.Clickable
var printTestBtns []widget.Clickable

// scrollovanie
var examList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

// Exams renders the "Exams" tab with dynamically generated columns based on data from the database.
func Exams(gtx layout.Context, th *themeUI.Theme, selectedTestID *uint, db *gorm.DB, tm *tabmanager.TabManager) layout.Dimensions {
	//logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	headerSize := unit.Sp(17)
	insetWidth := unit.Dp(15)

	tests, err := repository.GetAllTests(db)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní testov", slog.String("error", err.Error()))
		return layout.Dimensions{}
	}

	columns := []string{"Názov", "Rok", "Počet otázok", "Počet študentov", "Dátum", "Ukázať odpovede", "Vymazať", "Vyhodnotiť", "Tlačiť"}
	columnWidths := []float32{0.2, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1} // Pomery šírok
	if len(deleteButtons) != len(tests) {
		deleteButtons = make([]widget.Clickable, len(tests))
	}
	if len(showAnsButtons) != len(tests) {
		showAnsButtons = make([]widget.Clickable, len(tests))
	}
	if len(evaluateTestBtns) != len(tests) {
		evaluateTestBtns = make([]widget.Clickable, len(tests))
	}
	if len(printTestBtns) != len(tests) {
		printTestBtns = make([]widget.Clickable, len(tests))
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: insetWidth, Right: insetWidth}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.List(th.Theme, &examList).Layout(gtx, len(tests), func(gtx layout.Context, i int) layout.Dimensions {
					if i == 0 { // Prvá položka je hlavička
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[0])						
							}),
							layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[1])
							}),
							layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[2])
							}),
							layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[3])
							}),
							layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[4])
							}),
							layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[5])
							}),
							layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[6])
							}),
							layout.Flexed(columnWidths[7], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[7])
							}),
							layout.Flexed(columnWidths[8], func(gtx layout.Context) layout.Dimensions {
								return widgets.LabelBorder(gtx, th, headerSize, columns[8])
							}),
						)
					}
					test := tests[i-1]
					if deleteButtons[i].Clicked(gtx) {
						deleteTest(test.ID, db)
						tests = removeTestFromList(tests, i) // Remove test from the list for UI update
					}
					if showAnsButtons[i].Clicked(gtx) {
						showAnsTest(test.ID)

					}
					if evaluateTestBtns[i].Clicked(gtx) {
						*selectedTestID = test.ID // Nastavenie ID testu
						tm.ActiveTab = 3          // Prechod na UploadTab

					}
					if printTestBtns[i-1].Clicked(gtx) {
						printTest(test.ID)

					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
							return widgets.Body1Border(gtx, th, test.Title)
						}),
						layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
							return widgets.Body1Border(gtx, th, test.SchoolYear)
						}),
						layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
							return widgets.Body1Border(gtx, th, fmt.Sprintf("%d", test.QuestionCount))
						}),
						layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
							return widgets.Body1Border(gtx, th, fmt.Sprintf("%d", len(test.Students)))
						}),
						layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
							return widgets.Body1Border(gtx, th, "datum")
						}),
						layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
							btn := widgets.Button(th.Theme, &showAnsButtons[i], widgets.SearchIcon, widgets.IconPositionStart, "Zobraziť")
							btn.Background = themeUI.LightBlue
							btn.Color = themeUI.White
							return btn.Layout(gtx, th)
						}),
						layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
							btn := widgets.Button(th.Theme, &deleteButtons[i], widgets.DeleteIcon, widgets.IconPositionStart, "Vymazať")
							btn.Background = themeUI.Red
							btn.Color = themeUI.White
							return btn.Layout(gtx, th)
						}),
						layout.Flexed(columnWidths[7], func(gtx layout.Context) layout.Dimensions {
							btn := widgets.Button(th.Theme, &evaluateTestBtns[i], widgets.UploadIcon, widgets.IconPositionStart, "Vyhodnotiť")
							btn.Background = themeUI.LightGreen
							btn.Color = themeUI.White
							return btn.Layout(gtx, th)
						}),
						layout.Flexed(columnWidths[8], func(gtx layout.Context) layout.Dimensions {
							btn := widgets.Button(th.Theme, &printTestBtns[i], widgets.SaveIcon, widgets.IconPositionStart, "Tlačiť")
							btn.Background = themeUI.Gray
							btn.Color = themeUI.White
							return btn.Layout(gtx, th)
						}),
					)
				})
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

func printTest(Id uint) {
	logger := logging.GetLogger()

	logger.Info("tlačenie testu s ID", slog.Uint64("ID", uint64(Id)))
}
