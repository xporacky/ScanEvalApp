package tabs

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/io/system"
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
		go func() {
			// Získanie vybraného súboru (len jeden)
			file, err := t.explorer.ChooseFile() // Používame ChooseFile na výber jedného súboru
			if err != nil {
				log.Println("Chyba pri výbere súboru:", err)
				return
			}

			// Ak je súbor vybraný, načítame ho
			if file != nil {
				// Ak súbor implementuje io.Reader, môžeme ho prečítať
				if b, err := io.ReadAll(file); err == nil {
					t.selectedFile = string(b) // Uloženie obsahu súboru do premennej
				} else {
					log.Println("Chyba pri čítaní súboru:", err)
				}
			}
		}()
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


//BIG PROBLEM!!!
func (t *UploadTab) HandleEvent(event system.Event) {
	switch e := event.(type) {
	case system.FrameEvent:
		// Počúvanie udalostí pre Explorer
		t.explorer.ListenEvents(e)
	}
}

// Funkcia Upload pre vykreslenie záložky "Upload"
func Upload(gtx layout.Context, th *material.Theme, w *app.Window) layout.Dimensions {
	// Vytvorenie nového UploadTab
	uploadTab := NewUploadTab(w)

	// Spracovanie udalostí
	for _, event := range w.Events() { // Získame všetky udalosti z okna
		uploadTab.HandleEvent(event)
	}

	// Vykreslenie obsahu záložky
	return uploadTab.Layout(gtx, th)
}
