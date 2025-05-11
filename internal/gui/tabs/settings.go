package tabs

import (
	"log"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"ScanEvalApp/internal/config"
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"

	"github.com/sqweek/dialog"
)

type SettingTab struct {
	selectFolderBtn widget.Clickable
	selectedPath    string
}

func NewSettingTab(w *app.Window) *SettingTab {
	tab := &SettingTab{
		selectFolderBtn: widget.Clickable{},
	}

	if path, err := config.LoadLastPath(); err == nil {
		tab.selectedPath = path
	} else {
		log.Println("Nepodarilo sa načítať poslednú cestu:", err)
	}

	return tab
}

func (t *SettingTab) Layout(gtx layout.Context, th *themeUI.Theme, w *app.Window) layout.Dimensions {
	if t.selectFolderBtn.Clicked(gtx) {
		go func() {
			dir, err := dialog.Directory().Title("Vyber priečinok").Browse()
			if err != nil {
				log.Println("Chyba pri výbere priečinka:", err)
				return
			}
			t.selectedPath = dir

			if err := config.SaveLastPath(dir); err != nil {
				log.Println("Chyba pri ukladaní cesty:", err)
			}

			w.Invalidate()
		}()
	}
	return layout.Inset{
		Top:    unit.Dp(16),
		Left:   unit.Dp(16),
		Right:  unit.Dp(16),
		Bottom: unit.Dp(0),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Horizontal,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:   unit.Dp(5),
							Right: unit.Dp(8),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Label(th.Theme, unit.Sp(18), "Miesto ukladania súborov:").Layout(gtx)
						})
					}),
					layout.Flexed(0.1, func(gtx layout.Context) layout.Dimensions {
						text := "Žiadny priečinok nebol vybraný"
						if t.selectedPath != "" {
							text = t.selectedPath
						}
						return layout.Inset{
							Right: unit.Dp(8),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return widgets.LabelBorder(gtx, th, unit.Sp(16), text)
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:   unit.Dp(2),
							Right: unit.Dp(700),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							btn := widgets.Button(th.Theme, &t.selectFolderBtn, widgets.FileFolderIcon, widgets.IconPositionStart, "Vybrať priečinok")
							btn.Background = themeUI.LightGreen
							btn.Color = themeUI.White
							return btn.Layout(gtx, th)
						})
					}),
				)
			}),
		)
	})

}
