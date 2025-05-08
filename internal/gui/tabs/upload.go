package tabs

import (
	"fmt"
	"strings"

	"gioui.org/app"

	//"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"

	//"io"
	"log"
	//"path/filepath"
	//"ScanEvalApp/internal/database/models"
	"ScanEvalApp/internal/common"
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/files"
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"
	"ScanEvalApp/internal/scanprocessing"

	"time"

	"gorm.io/gorm"

	//"encoding/csv"
	"os"
)

type UploadTab struct {
	button       widget.Clickable
	explorer     *explorer.Explorer
	filePath     string
	examID       uint
	progressChan chan string
	progressText string
	dropdown     Dropdown
	uploadModal  *widgets.Modal
}

type Dropdown struct {
	open      bool
	button    widget.Clickable
	options   []string
	selected  widget.Enum
	lastValue string
}

func NewUploadTab(w *app.Window) *UploadTab {
	tab := &UploadTab{
		explorer:     explorer.NewExplorer(w),
		progressChan: make(chan string, 100),
		dropdown:     NewDropdown(),
		uploadModal:  widgets.NewModal(),
	}
	go func() {
		for {
			time.Sleep(1 * time.Second)
			select {
			case msg := <-tab.progressChan:
				tab.progressText = msg
				w.Invalidate()
			default:
				w.Invalidate()
			}
		}
	}()
	return tab
}

func (t *UploadTab) SetTestID(id uint) {
	t.examID = id
}

func (t *UploadTab) Layout(gtx layout.Context, th *themeUI.Theme, db *gorm.DB, w *app.Window) layout.Dimensions {
	// Spracovanie kliknutí na tlačidlo
	if t.button.Clicked(gtx) {
		go t.openFileDialog(db, th)
	}

	if t.dropdown.button.Clicked(gtx) {
		t.dropdown.open = !t.dropdown.open
	}

	// Celý obsah obalíme do layout.Stack
	return layout.Stack{}.Layout(gtx,
		// Najskôr vykreslíme hlavnú obrazovku
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceEvenly,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Button(th.Theme, &t.button, "Vybrať súbor").Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := "Žiadny súbor nebol vybraný"
						if t.filePath != "" {
							text = fmt.Sprintf("Vybraný súbor: %s", t.filePath)
						}
						return material.Label(th.Theme, unit.Sp(16), text).Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := "Žiadny test nebol vybraný"
						if t.examID != 0 {
							text = fmt.Sprintf("Vybraný test ID: %d", t.examID)
						}
						return material.Label(th.Theme, unit.Sp(16), text).Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return t.dropdown.Layout(gtx, th.Theme)
					}),
				)
			})
		}),
		// Potom modal navrch (ak je viditeľný)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if t.uploadModal.Visible {
				return t.uploadModal.Layout(gtx, th)
			}
			return layout.Dimensions{}
		}),
	)
}

// Funkcia na otvorenie dialógového okna na výber súboru
func (t *UploadTab) openFileDialog(db *gorm.DB, th *themeUI.Theme) {
	file, err := t.explorer.ChooseFile()
	if err != nil {
		log.Println("Chyba pri výbere súboru:", err)
		return
	}
	if file != nil {
		defer file.Close()

		if f, ok := file.(*os.File); ok {
			t.filePath = f.Name()
			fmt.Println("Cesta k súboru:", t.filePath)
		} else {
			log.Println("file nie je typu *os.File")
		}
		t.uploadModal.Visible = true
		t.uploadModal.Content = t.BuildProgressContent(th)
		t.uploadModal.SetCloseBtnEnable = false
		scanProcess(t, db)
	}
}

func (t *UploadTab) BuildProgressContent(th *themeUI.Theme) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEvenly}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Label(th.Theme, unit.Sp(25), "Spracovanie PDF:").Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Label(th.Theme, unit.Sp(20), t.progressText).Layout(gtx)
					}),
				)
			})
		})
	}
}

// Spracovanie eventov (pre Explorer)
func (t *UploadTab) HandleEvent(evt interface{}) { // Zmena na interface{}
	switch e := evt.(type) {
	case app.FrameEvent:
		t.explorer.ListenEvents(e)
	}
}

func scanProcess(t *UploadTab, db *gorm.DB) {
	var counter int = 0
	if t.examID == 0 && t.filePath == "" {
		fmt.Println("nevybrané povinné súbory")
		return
	}

	exam, err := repository.GetExam(db, t.examID)
	if err != nil {
		t.progressChan <- "Chyba pri načítaní testu."
		return
	}

	hadFailures := false
	t.progressChan <- "Spracovanie PDF sa začalo..."
	scanprocessing.ProcessPDF(t.filePath, exam, db, t.progressChan, &counter, &hadFailures)

	safeTitle := strings.ReplaceAll(exam.Title, " ", "_")
	safeTitle = repository.RemoveDiacritics(safeTitle)

	if hadFailures {
		t.progressChan <- fmt.Sprintf("Niektoré strany sa nepodarilo spracovať\nPDF bolo uložené do: %s%s_%d_failed_pages.pdf", common.GLOBAL_EXPORT_DIR, safeTitle, t.examID)
	} else {
		t.progressChan <- "Spracovanie dokončené."
	}

	_, err = repository.GetExam(db, t.examID)
	if err != nil {
		t.progressChan <- "Chyba pri získavaní údajov po spracovaní."
		return
	}

	t.uploadModal.SetCloseBtnEnable = true
}

func NewDropdown() Dropdown {
	options, err := files.GetFilesFromConfigs()
	if err != nil {
		log.Println("Error reading config files:", err)
	}

	dropdown := Dropdown{
		options:  options,
		selected: widget.Enum{},
	}

	if len(options) > 0 {
		dropdown.selected.Value = options[0]
		if err := scanprocessing.LoadConfig(options[0]); err != nil {
			log.Println("Chyba pri načítaní predvoleného konfigu:", err)
		}

	}

	return dropdown
}

func (d *Dropdown) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Ak sa zmenil výber, načítaj konfiguráciu
	if d.selected.Value != d.lastValue {
		d.lastValue = d.selected.Value
		err := scanprocessing.LoadConfig(d.selected.Value)
		if err != nil {
			fmt.Println("Chyba pri načítaní konfiguračného súboru:", err)
		}
	}

	children := make([]layout.FlexChild, len(d.options))
	for i, opt := range d.options {
		option := opt
		children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			radio := material.RadioButton(th, &d.selected, option, option)
			return radio.Layout(gtx)
		})
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx, children...)
}
