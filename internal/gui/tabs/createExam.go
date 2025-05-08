package tabs

/*
TODO:
dorobit ulozenie do databazy
dorobit scroll pri generovani otazok
povolit iba jednu moznost pri danej otazke
osetrit vstupy
*/
import (
	"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files/csv"
	"ScanEvalApp/internal/gui/tabmanager"
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"
	"ScanEvalApp/internal/logging"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"gorm.io/gorm"
)

var (
	nameInput      widget.Editor
	schoolYear     widget.Editor
	datetimeInput  widget.Editor
	questionsInput widget.Editor
	submitButton   widget.Clickable
	createButton   widget.Clickable
	questionForms  []questionForm
	showQuestions  bool
)

type UploadCsv struct {
	button       widget.Clickable
	explorer     *explorer.Explorer
	selectedFile string
	filePath     string
}

func NewUploadCsv(w *app.Window) *UploadCsv {
	return &UploadCsv{
		explorer: explorer.NewExplorer(w),
	}
}

var questionList widget.List = widget.List{List: layout.List{Axis: layout.Vertical}}

type questionForm struct {
	selectedOption widget.Enum // Uchováva vybranú možnosť (A, B, C, D, E)
}

// CreateExam renders the content for the "Vytvorenie Písomky" tab.
func (t *UploadCsv) CreateExam(gtx layout.Context, th *themeUI.Theme, db *gorm.DB, tm *tabmanager.TabManager) layout.Dimensions {
	logger := logging.GetLogger()
	columnWidths := []float32{0.4, 0.2, 0.2, 0.2}
	insetwidth := unit.Dp(10)
	if createButton.Clicked(gtx) {
		logger.Info("Kliknutie na tlačidlo Vytvoriť test")
		if questionsInput.Text() != "" {
			n := parseNumber(questionsInput.Text())
			if n > 0 {
				updateQuestionForms(n)
				showQuestions = true
			}
		}
	}
	if t.button.Clicked(gtx) {
		go t.openFileDialog(db)
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 4, Bottom: 2}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(columnWidths[0], func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							editor := widgets.NewEditorField(th.Theme, &nameInput, "Názov") // Šírku riadi columnWidths
							return editor.Layout(gtx, th)
						})
					}),
					layout.Flexed(columnWidths[1], func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							editor := widgets.NewEditorField(th.Theme, &schoolYear, "Šk. rok (YYYY/YY)")
							return editor.Layout(gtx, th)
						})
					}),
					layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							editor := widgets.NewEditorField(th.Theme, &datetimeInput, "Dátum a Čas (dd.MM.yyyy HH:mm)")
							return editor.Layout(gtx, th)
						})
					}),
					layout.Flexed(columnWidths[3], func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							editor := widgets.NewEditorField(th.Theme, &questionsInput, "Počet otázok")
							return editor.Layout(gtx, th)
						})
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
				layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := widgets.Button(th.Theme, &t.button, widgets.FileFolderIcon, widgets.IconPositionStart, "Nahrať študentov (.csv)")
						btn.Background = themeUI.LightYellow
						btn.Color = themeUI.Black
						return btn.Layout(gtx, th)
					})
				}),
				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						text := "Žiadny súbor nebol vybraný"
						if t.selectedFile != "" {
							text = fmt.Sprintf("Vybraný súbor: %s", t.filePath)
						}
						return material.Label(th.Theme, unit.Sp(16), text).Layout(gtx)
					})
				}),
				layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := widgets.Button(th.Theme, &createButton, widgets.PlusIcon, widgets.IconPositionStart, "Generovať test")
						btn.Background = themeUI.LightBlue
						btn.Color = themeUI.White
						return btn.Layout(gtx, th)
					})
				}),
			)
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if showQuestions {
				return material.List(th.Theme, &questionList).Layout(gtx, len(questionForms)+1, func(gtx layout.Context, i int) layout.Dimensions {
					if i < len(questionForms) {
						qf := &questionForms[i]
						return layout.Flex{
							Axis:    layout.Horizontal,
							Spacing: layout.SpaceAround,
						}.Layout(gtx, renderOptions(gtx, th, i+1, qf)...)
					}

					// Posledný element - tlačidlo "Vytvoriť test"
					return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := widgets.Button(th.Theme, &submitButton, widgets.SaveIcon, widgets.IconPositionStart, "Vytvoriť test")
						btn.Background = themeUI.LightGreen
						btn.Color = themeUI.Black
						if submitButton.Clicked(gtx) {
							logger.Info("Vybraný súbor", slog.String("Path", t.filePath))
							logger.Info("Kliknutie na tlačidlo Odoslať")
							submitForm(db, t, tm)
						}
						return btn.Layout(gtx, th)
					})
				})
			}
			return layout.Dimensions{}
		}),
	)
}

// Funkcia na otvorenie dialógového okna na výber súboru
func (t *UploadCsv) openFileDialog(db *gorm.DB) {
	errorLogger := logging.GetErrorLogger()

	file, err := t.explorer.ChooseFile()
	if err != nil {
		errorLogger.Error("Chyba pri výbere súboru:", slog.String("error", err.Error()))
		return
	}
	if file != nil {
		defer file.Close() // Nezabudni zatvoriť súbor
		if f, ok := file.(*os.File); ok {
			t.filePath = f.Name()
		} else {
			errorLogger.Error("File nie je typu *os.File")
		}
		b, err := io.ReadAll(file)
		if err != nil {
			errorLogger.Error("Chyba pri čítaní súboru:", slog.String("error", err.Error()))
			return
		}
		t.selectedFile = string(b)
	}
}

// Spracovanie eventov (pre Explorer)
func (t *UploadCsv) HandleEvent(evt interface{}) { // Zmena na interface{}
	switch e := evt.(type) {
	case app.FrameEvent:
		t.explorer.ListenEvents(e)
	}
}

func parseNumber(input string) int {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	num, err := strconv.Atoi(input)
	if err != nil {
		errorLogger.Error("Chyba pri parsovaní počtu otázok", slog.String("error", err.Error()))
	} else {
		logger.Debug("Preparsované číslo pocet otazok:", slog.Int("count", num))
	}
	return num
}

func updateQuestionForms(n int) {
	for len(questionForms) < n {
		questionForms = append(questionForms, questionForm{})
	}
	for len(questionForms) > n {
		questionForms = questionForms[:len(questionForms)-1]
	}
}

func renderQuestionForms(gtx layout.Context, th *themeUI.Theme) []layout.FlexChild {
	children := make([]layout.FlexChild, len(questionForms))
	for i := range questionForms { // Prechádzame len indexy, aby sme pracovali priamo so slice-om
		qf := &questionForms[i] // Uložíme si pointer na konkrétny prvok
		children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceAround,
			}.Layout(gtx, renderOptions(gtx, th, i+1, qf)...) // Odovzdávame pointer na správny prvok
		})
	}
	return children
}

func renderOptions(gtx layout.Context, th *themeUI.Theme, questionIndex int, qf *questionForm) []layout.FlexChild {
	options := []string{"A", "B", "C", "D", "E"}
	children := make([]layout.FlexChild, len(options)+1) // Prvý prvok je číslo otázky

	// Pridáme číslo otázky (napr. "01:")
	children[0] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return material.Label(th.Theme, unit.Sp(15), fmt.Sprintf("%02d:", questionIndex)).Layout(gtx)
	})

	// Vykreslíme rádio tlačidlá pre možnosti A–E
	for i, option := range options {
		i, option := i, option
		children[i+1] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.RadioButton(th.Theme, &qf.selectedOption, option, option).Layout(gtx)
		})
	}

	return children
}
func isValidSchoolYear(schoolYear string) bool {
	re := regexp.MustCompile(`^\d{4}/\d{2}$`)
	return re.MatchString(schoolYear)
}

func parseDateTime(dateTime string) (time.Time, bool) {
	parsedTime, err := time.Parse("02.01.2006 15:04", dateTime)
	if err != nil {
		return time.Time{}, false
	}
	return parsedTime, true
}

func submitForm(db *gorm.DB, t *UploadCsv, tm *tabmanager.TabManager) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	nazov := nameInput.Text()
	skrok := schoolYear.Text()
	if !isValidSchoolYear(skrok) {
		// Použijeme logger na logovanie chyby
		errorLogger.Error("Neplatný školský rok", slog.Group("INFO", slog.String("sk.rok", skrok)))
		return
	}
	datumacas := datetimeInput.Text()
	parsedDateTime, valid := parseDateTime(datumacas)
	if !valid {
		// Logovanie chyby s detailmi
		errorLogger.Error("Neplatný dátum a čas", slog.Group("INFO", slog.String("datumacas", datumacas)))
		return
	}
	pocetOtazok, err := strconv.Atoi(questionsInput.Text())
	if err != nil {
		errorLogger.Error("Chyba pri parsovaní počtu otázok", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return
	}

	var answers []string
	for _, qf := range questionForms {
		answers = append(answers, qf.selectedOption.Value)
	}
	answersStr := strings.Join(answers, "")

	// Vytvorenie testu
	exam := models.Exam{
		Title:         nazov,
		SchoolYear:    skrok,
		Date:          parsedDateTime,
		QuestionCount: pocetOtazok,
		Questions:     answersStr,
	}
	// ulozenie do db
	err = repository.CreateExam(db, &exam)
	if err != nil {
		errorLogger.Error("Chyba pri ukladaní testu", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return
	}

	err = csv.ImportStudentsFromCSV(db, t.selectedFile, exam.ID)
	if err != nil {
		errorLogger.Error("Chyba pri importe študentov z CSV", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return
	}

	logger.Info("Test bol úspešne vytvorený",
		slog.String("examTitle", exam.Title),
		slog.String("examID", strconv.Itoa(int(exam.ID))),
		slog.Int("questionCount", exam.QuestionCount))

	// Resetovanie vstupov
	nameInput.SetText("")
	schoolYear.SetText("")
	datetimeInput.SetText("")
	questionsInput.SetText("")
	t.selectedFile = ""
	t.filePath = ""
	showQuestions = false

	// Resetovanie otázok
	questionForms = nil
	tm.ActiveTab = 0
}
