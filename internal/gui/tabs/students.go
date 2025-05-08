package tabs

import (
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files/pdf"
	"fmt"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gorm.io/gorm"

	//"image/color"

	//"reflect"
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"
	"ScanEvalApp/internal/logging"
	"log/slog"

	"ScanEvalApp/internal/latex"

	"gioui.org/unit"
)

var printButtons []widget.Clickable
var downloadButtons []widget.Clickable
var searchQuery widget.Editor
var studentModal widgets.Modal

// scrollovanie
var studentList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

// StudentsTab renders the "Students" tab with a table of students.
func Students(gtx layout.Context, th *themeUI.Theme, db *gorm.DB) layout.Dimensions {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()
	insetWidth := unit.Dp(15)
	headerSize := unit.Sp(17)

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
	columns := []string{"Meno", "Priezvisko", "Dátum narodenia", "Registračné číslo", "Miestnosť", "Skóre", "Tlačiť hárok", "Stiahnúť hárok"}
	columnWidths := []float32{0.2, 0.2, 0.15, 0.15, 0.1, 0.05, 0.075, 0.075} // Pomery šírok
	if len(printButtons) != len(students) {
		printButtons = make([]widget.Clickable, len(students))
	}
	if len(downloadButtons) != len(students) {
		downloadButtons = make([]widget.Clickable, len(students))
	}
	return layout.Stack{}.Layout(gtx,
		// Hlavný obsah aplikácie
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						editor := widgets.NewEditorField(th.Theme, &searchQuery, "🔎   Vyhľadávanie (Meno, Priezvisko, Registračné číslo)") // Šírku riadi columnWidths
						return editor.Layout(gtx, th)
					})
				}),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: insetWidth, Right: insetWidth}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.List(th.Material(), &studentList).Layout(gtx, len(students)+1, func(gtx layout.Context, i int) layout.Dimensions {
							if i == 0 {
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
									layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
										return widgets.LabelBorder(gtx, th, headerSize, columns[7])
									}),
								)
							}

							student := students[i-1]
							if printButtons[i-1].Clicked(gtx) {
								// iba na testovanie, ci pocita dobre score
								fmt.Println("Student: ", student.RegistrationNumber, "Odpovede: ", student.Answers, ", score: ", student.Score)
								//printSheet(student.RegistrationNumber)
								studentModal.Visible = true
								studentModal.SetCloseBtnEnable = false
								isGenerating := true
								generatedPath := ""
								studentModal.Content = widgets.ContentGenerating(th, &isGenerating, &generatedPath)

								go func() {
									path, err := latex.PrintSheet(db, student.RegistrationNumber)
									if err != nil {
										errorLogger.Error("Chyba pri tlači hárku pre študenta",
											"student_id", student.ID,
											slog.Uint64("registration_number", uint64(student.RegistrationNumber)),
											slog.String("path", path),
											slog.String("error", err.Error()))
									} else {
										generatedPath = path
										isGenerating = false
										studentModal.SetCloseBtnEnable = true
										logger.Info("Úspešne vytlačený hárok pre študenta",
											slog.Uint64("registration_number", uint64(student.RegistrationNumber)))
									}
								}()
							}
							if downloadButtons[i-1].Clicked(gtx) {
								fmt.Printf("stiahnuť vyplneny harok")
								// sem si zavolam funkciu, ktora pre studenta slicne z pdf dane subory a to ulozi ako pdf do tmp s nazvom studentovho id
								err := pdf.SlicePdfForStudent(db, student.RegistrationNumber)
								if err != nil {
									errorLogger.Error("Chyba pri slicingu PDF pre študenta", "registration_number", student.RegistrationNumber, "error", err.Error())
								} else {
									logger.Info("Úspešne slicitované PDF pre študenta", "registration_number", student.RegistrationNumber)
								}
							}

							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, student.Name)
								}),
								layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, student.Surname)
								}),
								layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, student.BirthDate.Format("2006-01-02"))
								}),
								layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, fmt.Sprintf("%d", student.RegistrationNumber))
								}),
								layout.Flexed(columnWidths[4], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, student.Room)
								}),
								layout.Flexed(columnWidths[5], func(gtx layout.Context) layout.Dimensions {
									return widgets.Body1Border(gtx, th, fmt.Sprintf("%d", student.Score))
								}),
								layout.Flexed(columnWidths[6], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &printButtons[i-1], widgets.SaveIcon, widgets.IconPositionStart, "Tlačiť")
									btn.Background = themeUI.Gray
									btn.Color = themeUI.White
									return btn.Layout(gtx, th)
								}),
								layout.Flexed(columnWidths[7], func(gtx layout.Context) layout.Dimensions {
									btn := widgets.Button(th.Theme, &downloadButtons[i-1], widgets.SaveIcon, widgets.IconPositionStart, "Stiahnúť")
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
			if studentModal.Visible {
				return studentModal.Layout(gtx, th)
			}
			return layout.Dimensions{}
		}),
	)
}

func printSheet(registrationNumber string) {
	logger := logging.GetLogger()

	logger.Info("Volám tlač hárku pre študenta ID", slog.String("ID", registrationNumber))
}
