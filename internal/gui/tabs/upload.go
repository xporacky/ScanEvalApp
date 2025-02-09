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
	"io"
	"log"
)

type UploadTab struct {
	button       widget.Clickable
	explorer     *explorer.Explorer
	selectedFile string
}

func NewUploadTab(w *app.Window) *UploadTab {
	return &UploadTab{
		explorer: explorer.NewExplorer(w),
	}
}

func (t *UploadTab) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Spracovanie kliknutí na tlačidlo
	if t.button.Clicked(gtx) {
		go t.openFileDialog()
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
				if t.selectedFile != "" {
					text = fmt.Sprintf("Vybraný súbor: %s", t.selectedFile)
				}
				return material.Label(th, unit.Sp(16), text).Layout(gtx)
			}),
		)
	})
}

// Funkcia na otvorenie dialógového okna na výber súboru
func (t *UploadTab) openFileDialog() {
	file, err := t.explorer.ChooseFile()
	if err != nil {
		log.Println("Chyba pri výbere súboru:", err)
		return
	}
	if file != nil {
		defer file.Close() // Nezabudni zatvoriť súbor
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println("Chyba pri čítaní súboru:", err)
			return
		}
		t.selectedFile = string(b)
	}
}

// Spracovanie eventov (pre Explorer)
func (t *UploadTab) HandleEvent(evt interface{}) { // Zmena na interface{}
	switch e := evt.(type) {
	case app.FrameEvent:
		t.explorer.ListenEvents(e)
	}
}
