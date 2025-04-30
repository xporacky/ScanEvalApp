package tabs

import (
	"ScanEvalApp/internal/csvhelper"
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/latex"

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
var showGenStatButtons []widget.Clickable
var evaluateExamBtns []widget.Clickable
var printExamBtns []widget.Clickable
var exportCSV []widget.Clickable
var modal widgets.Modal

// scrollovanie
var examList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

// Exams renders the "Exams" tab with dynamically generated columns based on data from the database.

func Exams(gtx layout.Context, th *themeUI.Theme, selectedExamID *uint, db *gorm.DB, tm *tabmanager.TabManager) layout.Dimensions {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	headerSize := unit.Sp(17)
	insetWidth := unit.Dp(15)

	exams, err := repository.GetAllExams(db)
	if err != nil {
		errorLogger.Error("Chyba pri načítaní testov", slog.String("error", err.Error()))
		return layout.Dimensions{}
	}

	columns := []string{"Názov", "Rok", "Počet otázok", "Počet študentov", "Dátum", "Ukázať odpovede", "Štatistika", "Vymazať", "Vyhodnotiť", "Tlačiť", "CSV"}
	columnWidths := []float32{0.180, 0.06, 0.08, 0.1, 0.075, 0.1, 0.1, 0.1, 0.1, 0.1, 0.05} // Pomery šírok
	if len(deleteButtons) != len(exams) {
		deleteButtons = make([]widget.Clickable, len(exams))
	}
	if len(showAnsButtons) != len(exams) {
		showAnsButtons = make([]widget.Clickable, len(exams))
	}
	if len(showGenStatButtons) != len(exams) {
		showGenStatButtons = make([]widget.Clickable, len(exams))
	}
	if len(evaluateExamBtns) != len(exams) {
		evaluateExamBtns = make([]widget.Clickable, len(exams))
	}
	if len(printExamBtns) != len(exams) {
		printExamBtns = make([]widget.Clickable, len(exams))
	}
	if len(exportCSV) != len(exams) {
		exportCSV = make([]widget.Clickable, len(exams))
	}
	return layout.Stack{}.Layout(gtx,
		// Hlavný obsah aplikácie
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: insetWidth, Right: insetWidth}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.List(th.Theme, &examList).Layout(gtx, len(exams)+1, func(gtx layout.Context, i int) layout.Dimensions {
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
									layout.Flexed(columnWidths[9], func(gtx layout.Context) layout.Dimensions {
										return widgets.LabelBorder(gtx, th, headerSize, columns[9])
									}),
									layout.Flexed(columnWidths[10], func(gtx layout.Context) layout.Dimensions {
										return widgets.LabelBorder(gtx, th, headerSize, columns[10])
									}),
								)
							}
							exam := exams[i-1]
							if deleteButtons[i-1].Clicked(gtx) {
								deleteExam(db, &exam)
								exams = removeExamFromList(exams, i-1) // Remove exam from the list for UI update
							}
							if showAnsButtons[i-1].Clicked(gtx) {
								modal.Visible = true
								modal.SetCloseBtnEnable = true
								modal.Content = BuildAnswersContent(exam.Questions, th)
							}
							if showGenStatButtons[i-1].Clicked(gtx) {
								modal.Visible = true
								modal.SetCloseBtnEnable = true
								modal.Content = Statistics(gtx, th, &exam)
							}
							if evaluateExamBtns[i-1].Clicked(gtx) {
								*selectedExamID = exam.ID // Nastavenie ID testu
								tm.ActiveTab = 3          // Prechod na UploadTab

							}
							if printExamBtns[i-1].Clicked(gtx) {
								modal.Visible = true
								modal.SetCloseBtnEnable = false
								isGenerating := true
								generatedPath := ""
								modal.Content = widgets.ContentGenerating(th, &isGenerating, &generatedPath)
								go func() {
									path, err := latex.ParallelGeneratePDFs(db, exam.ID, latex.TEMPLATE_PATH, latex.OUTPUT_PDF_PATH)

									if err != nil {
										errorLogger.Error("Chyba pri generovaní PDF",
											slog.String("error", err.Error()),
											slog.String("path", path),
										)
									} else {
										generatedPath = path
										isGenerating = false
										modal.SetCloseBtnEnable = true
										logger.Info("Úspešne vygenerované PDF pre skúšku", slog.String("examID", fmt.Sprintf("%d", exam.ID)))
									}
								}()
							}
							if exportCSV[i-1].Clicked(gtx) {
								modal.Visible = true
								modal.SetCloseBtnEnable = false
								isGenerating := true
								generatedPath := ""
								modal.Content = widgets.ContentGenerating(th, &isGenerating, &generatedPath)
								go func() {
									path, err := csvhelper.ExportStudentsToCSV(db, exam)

									if err != nil {
										errorLogger.Error("Chyba pri exportovani CSV", slog.String("error", err.Error()))
									} else {
										generatedPath = path
										isGenerating = false
										modal.SetCloseBtnEnable = true
										logger.Info("Úspešne vyexportovane CSV pre skúšku", slog.String("examID", fmt.Sprintf("%d", exam.ID)))
									}
								}()
							}

							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, exam.Title)
								}),
								layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, exam.SchoolYear)
								}),
								layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, fmt.Sprintf("%d", exam.QuestionCount))
								}),
								layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, fmt.Sprintf("%d", len(exam.Students)))
								}),
								layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, exam.Date.Format("02.01.2006 15:04"))
								}),
								layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &showAnsButtons[i-1], widgets.SearchIcon, widgets.IconPositionStart, "Zobraziť")
									btn.Background = themeUI.LightBlue
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
								layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &showGenStatButtons[i-1], widgets.MenuIcon, widgets.IconPositionStart, "Štatistika")
									btn.Background = themeUI.LightGreen
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
								layout.Flexed(columnWidths[7], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &deleteButtons[i-1], widgets.DeleteIcon, widgets.IconPositionStart, "Vymazať")
									btn.Background = themeUI.Red
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
								layout.Flexed(columnWidths[8], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &evaluateExamBtns[i-1], widgets.UploadIcon, widgets.IconPositionStart, "Vyhodnotiť")
									btn.Background = themeUI.LightPurple
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
								layout.Flexed(columnWidths[9], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &printExamBtns[i-1], widgets.SaveIcon, widgets.IconPositionStart, "Tlačiť")
									btn.Background = themeUI.Green
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
								layout.Flexed(columnWidths[10], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &exportCSV[i-1], widgets.FileFolderIcon, widgets.IconPositionStart, "csv")
									btn.Background = themeUI.Gray
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
							)
						})
					})
				}),
			)
		}),
		// Modal - vykreslí sa NAVRCHU, ak je viditeľný
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if modal.Visible {
				return modal.Layout(gtx, th)
			}
			return layout.Dimensions{}
		}),
	)
}

func deleteExam(db *gorm.DB, exam *models.Exam) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Deleting the test from the database
	if err := repository.DeleteExam(db, exam); err != nil {
		errorLogger.Error("Chyba pri vymazávaní testu", slog.Uint64("ID", uint64(exam.ID)), slog.String("error", err.Error()))
		return
	}

	logger.Info("Vymazanie testu s ID", slog.Uint64("ID", uint64(exam.ID)))
}

func removeExamFromList(exams []models.Exam, index int) []models.Exam {
	// Removing the test from the list at the specified index
	return append(exams[:index], exams[index+1:]...)

}

func BuildAnswersContent(answers string, theme *themeUI.Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Label(theme.Material(), unit.Sp(20), "Odpovede testu:").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					chars := []rune(answers)
					rows := []layout.FlexChild{}
					for i := 0; i < len(chars); i += 10 {
						row := []layout.FlexChild{}
						for j := 0; j < 10 && i+j < len(chars); j++ {
							index := i + j + 1
							row = append(row, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return material.Label(theme.Material(), unit.Sp(16), fmt.Sprintf("%2d:   %c", index, chars[i+j])).Layout(gtx)
								})
							}))
						}
						rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, row...)
						}))
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
				}),
			)
		})
	}
}

var statisticsOptions []string = []string{
	"Maximum bodov",
	"Minimum bodov",
	"Priemer",
	"Medián",
	"Graf rozdelenia bodov celkovo",
	"Graf rozdelenia za jednotlivé príklady",
	"Úspešnosť absolútna aj relatívna",
	"Úspešnosť absolútna aj relatívna pre jednotlivé príklady",
}

var generateStatsButton widget.Clickable
var checkboxes []widget.Bool

func Statistics(gtx layout.Context, th *themeUI.Theme, exam *models.Exam) layout.Widget {
	errorLogger := logging.GetErrorLogger()

	checkboxes = make([]widget.Bool, len(statisticsOptions))
	// Nastaviť všetky checkboxy na zaškrtnuté
	for i := range checkboxes {
		checkboxes[i] = widget.Bool{Value: true}
	}
	return func(gtx layout.Context) layout.Dimensions {
		if generateStatsButton.Clicked(gtx) {
			modal.Visible = true
			modal.SetCloseBtnEnable = false
			isGenerating := true
			generatedPath := ""
			modal.Content = widgets.ContentGenerating(th, &isGenerating, &generatedPath)
			go func() {
				selectedStats := collectSelectedStats()
				path, err := latex.GenerateStatistics(selectedStats, exam)
				if err != nil {
					errorLogger.Error("Chyba pri generovaní štatistík", slog.String("error", err.Error()))
					isGenerating = false
					modal.SetCloseBtnEnable = true
					return
				}
				generatedPath = path
				isGenerating = false
				modal.SetCloseBtnEnable = true
			}()
		}
		return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Label(th.Material(), unit.Sp(20), "Vyberte požadované štatistiky:").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					var checkboxItems []layout.FlexChild
					for i, option := range statisticsOptions {
						checkboxItems = append(checkboxItems, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return widgets.Checkbox(gtx, th, &checkboxes[i], option, unit.Sp(15))
						}))
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, checkboxItems...)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := widgets.Button(th.Theme, &generateStatsButton, widgets.SaveIcon, widgets.IconPositionStart, "Generovať štatistiky")
					btn.Background = themeUI.LightGreen
					btn.Color = themeUI.White
					return btn.Layout(gtx, th)
				}),
			)
		})
	}
}

func collectSelectedStats() []string {
	var selected []string
	for i, checked := range checkboxes {
		if checked.Value {
			selected = append(selected, statisticsOptions[i])
		}
	}
	return selected
}
