package tabs
/*
TODO: 
dorobit ulozenie do databazy
dorobit scroll pri generovani otazok 
povolit iba jednu moznost pri danej otazke
osetrit vstupy
*/
import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/unit"
	"gioui.org/x/explorer"
	"io"
	"fmt"
	"strconv"
	"ScanEvalApp/internal/logging"
	"log/slog"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/database/models"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
	"encoding/csv"
	"os"
	"ScanEvalApp/internal/gui/themeUI"
	themeIU "ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"
)

var (
	nameInput      widget.Editor
	roomInput      widget.Editor
	timeInput      widget.Editor
	questionsInput widget.Editor
	submitButton   widget.Clickable
	createButton   widget.Clickable
	questionForms  []questionForm
	showQuestions  bool
)
type UploadCsv struct {
	button       	widget.Clickable
	explorer     	*explorer.Explorer
	selectedFile 	string
	filePath		string

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

// CreateTest renders the content for the "Vytvorenie Písomky" tab.
func (t *UploadCsv) CreateTest(gtx layout.Context, th *themeIU.Theme, db *gorm.DB) layout.Dimensions {
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
		Axis:    layout.Vertical,
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
								editor := widgets.NewEditorField(th.Theme, &roomInput, "Miestnosť")
								return editor.Layout(gtx, th)
							})
						}),
						layout.Flexed(columnWidths[2], func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(insetwidth).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								editor := widgets.NewEditorField(th.Theme, &timeInput, "Čas")
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
			btn := widgets.Button(th.Theme, &createButton, widgets.PlusIcon, widgets.IconPositionStart, "Vytvoriť test")
			btn.Background = themeUI.LightBlue
			btn.Color = themeUI.White
			if printAllButton.Clicked(gtx) {
				logger.Info("Kliknutie na tlačidlo Vytvoriť test")
				printAllSheets()
			}
			return btn.Layout(gtx, th)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := widgets.Button(th.Theme, &t.button, widgets.FileFolderIcon, widgets.IconPositionStart, "Nahrať študentov (.csv)")
			btn.Background = themeUI.LightYellow
			btn.Color = themeUI.Black
			if printAllButton.Clicked(gtx) {
				logger.Info("Kliknutie na tlačidlo nahrať študentov")
				printAllSheets()
			}
			return btn.Layout(gtx, th)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			text := "Žiadny súbor nebol vybraný"
			if t.selectedFile != "" {
				text = fmt.Sprintf("Vybraný súbor: %s", t.filePath)
				
				
			}
			return material.Label(th.Theme, unit.Sp(16), text).Layout(gtx)
		}),
		
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if showQuestions {
				btn := material.Button(th.Theme, &submitButton, "Odoslať")
				if submitButton.Clicked(gtx) {
					logger.Info("Kliknutie na tlačidlo Odoslať")
					submitForm(db, t)
				}
				return btn.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if showQuestions {
				return material.List(th.Theme, &questionList).Layout(gtx, len(questionForms), func(gtx layout.Context, i int) layout.Dimensions {
					qf := &questionForms[i]
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceAround,
					}.Layout(gtx, renderOptions(gtx, th, i+1, qf)...)
				})
			}
			return layout.Dimensions{}
		}),
	)
}

// Funkcia na otvorenie dialógového okna na výber súboru
func (t *UploadCsv) openFileDialog(db *gorm.DB) {
	file, err := t.explorer.ChooseFile()
	if err != nil {
		log.Println("Chyba pri výbere súboru:", err)
		return
	}
	if file != nil {
		defer file.Close() // Nezabudni zatvoriť súbor
		if f, ok := file.(*os.File); ok {
			t.filePath = f.Name()
		} else {
			log.Println("file nie je typu *os.File")
		}	
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println("Chyba pri čítaní súboru:", err)
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

func renderQuestionForms(gtx layout.Context, th *themeIU.Theme) []layout.FlexChild {
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



func renderOptions(gtx layout.Context,th *themeIU.Theme, questionIndex int, qf *questionForm) []layout.FlexChild {
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

func submitForm(db *gorm.DB, t *UploadCsv) {
	logger := logging.GetLogger()
	errorLogger := logging.GetErrorLogger()

	// Načítame údaje zo všetkých inputov
	nazov := nameInput.Text()
	miestnost := roomInput.Text()
	cas := timeInput.Text()
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
	test := models.Test{
		Title:      nazov,
		SchoolYear: cas,
//		Room:       miestnost,
		QuestionCount: pocetOtazok,
		Questions : answersStr,
	}
	// ulozenie do db
	err = repository.CreateTest(db, &test)
	if err != nil {
		errorLogger.Error("Chyba pri ukladaní testu", slog.Group("CRITICAL", slog.String("error", err.Error())))
		return
	}
	// Vytvoríme si výsledný výpis
	logger.Info("Formulár odoslaný", 
		slog.String("nazov", nazov), 
		slog.String("miestnost", miestnost), 
		slog.String("cas", cas), 
		slog.Int("pocetOtazok", pocetOtazok),
		slog.String("odpovede", answersStr))
		// Premenná pre uchovávanie zaškrtnutých možností

	fmt.Println(t.selectedFile)


	reader := csv.NewReader(strings.NewReader(t.selectedFile))
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Println("error1: %s", err)
		return 
	}


	fmt.Println("studenti v csv: %s", rows)

	for i, row := range rows {
		fmt.Println("som dnu for")
		if i == 0 {
			fmt.Println("hlavicka")
			continue // Preskočiť hlavičku CSV
		}
		birthDate, err := time.Parse("2006-01-02", row[2]) 
		if err != nil {
			fmt.Println("error2: %s", err)
			return
		}

		student := models.Student{
			Name:               row[0],
			Surname:            row[1],
			BirthDate:          birthDate,
			RegistrationNumber: row[3],
			Room:               row[4],
			TestID:             test.ID,					//TODO:preroobilt!!!!!!!!!!!!!!!!!!!!! 
		}
		if err := repository.CreateStudent(db, &student); err != nil {
			fmt.Println("error3: ", err)
			return 
		}else{
			fmt.Println("student pridany: %s", student)
		}
	}
}
