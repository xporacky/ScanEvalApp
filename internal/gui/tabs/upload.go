package tabs

import (
	"fmt"

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
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/scanprocessing"

	//"strings"
	//"time"
	"gorm.io/gorm"
	//"encoding/csv"
	"os"
)

type UploadTab struct {
	button   widget.Clickable
	explorer *explorer.Explorer
	filePath string
	examID   uint
	progressChan chan string
	progressText string
}

func (t *UploadTab) SetTestID(id uint) {
	t.examID = id
}

func NewUploadTab(w *app.Window) *UploadTab {
	return &UploadTab{
		explorer: explorer.NewExplorer(w),
		progressChan: make(chan string, 100),
	}
}

func (t *UploadTab) Layout(gtx layout.Context, th *material.Theme, db *gorm.DB) layout.Dimensions {
	// Spracovanie kliknutí na tlačidlo
	if t.button.Clicked(gtx) {
		go t.openFileDialog(db)
	}
	select {
	case msg := <-t.progressChan:
		t.progressText = msg
	default:
	}

	// Vykreslenie layoutu
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(th, &t.button, "Vybrať súbor").Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := "Žiadny súbor nebol vybraný"
				if t.filePath != "" {
					text = fmt.Sprintf("Vybraný súbor: %s", t.filePath)

				}
				return material.Label(th, unit.Sp(16), text).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := "Žiadny test nebol vybraný"
				if t.examID != 0 {
					text = fmt.Sprintf("Vybraný test ID: %d", t.examID)
				}
				return material.Label(th, unit.Sp(16), text).Layout(gtx)
			}),

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {				
				return material.Label(th, unit.Sp(16), t.progressText).Layout(gtx)
			}),
		)
	})
}

// Funkcia na otvorenie dialógového okna na výber súboru
func (t *UploadTab) openFileDialog(db *gorm.DB) {
	file, err := t.explorer.ChooseFile()
	if err != nil {
		log.Println("Chyba pri výbere súboru:", err)
		return
	}
	if file != nil {
		defer file.Close() // Nezabudni zatvoriť súbor

		// Pretypovanie na *os.File
		if f, ok := file.(*os.File); ok {
			t.filePath = f.Name()
			fmt.Println("Cesta k súboru:", t.filePath)
		} else {
			log.Println("file nie je typu *os.File")
		}

		scanProcess(t, db)
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
	if t.examID == 0 && t.filePath == "" {
		fmt.Println("nevybrane povinne subory")
		return
	}

	exam, err := repository.GetExam(db, t.examID)
	if err != nil {
		return
	}
	t.progressChan <- "Spracovanie PDF sa začalo..."
	scanprocessing.ProcessPDF(t.filePath, exam, db, t.progressChan)
	t.progressChan <- "Spracovanie Dokončene"
}
