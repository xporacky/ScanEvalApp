// window.go
package window

import (
	"ScanEvalApp/internal/gui/fonts"
	"ScanEvalApp/internal/gui/tabmanager"
	"ScanEvalApp/internal/gui/tabs"
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/logging"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
	"gorm.io/gorm"
)

func RunWindow(db *gorm.DB) error {
	logger := logging.GetLogger()

	w := new(app.Window)
	w.Option(
		app.Title("ScanEvalApp"),
		app.Maximized.Option(),
	)

	var ops op.Ops
	tm := tabmanager.NewTabManager(5)
	tabNames := []string{"Písomky", "Študenti", "Vytvorenie Písomky", "Vyhodnotenie testu", "Nastavenia"}

	uploadTab := tabs.NewUploadTab(w)
	settingTab := tabs.NewSettingTab(w)
	uploadCsv := tabs.NewUploadCsv(w)
	var selectedTestID uint

	for {
		evt := w.Event()
		switch typ := evt.(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, typ)
			ops.Reset()
			fontCollection, err := fonts.Prepare()
			if err != nil {
				return err
			}
			theme := material.NewTheme()
			theme.Shaper = text.NewShaper(text.WithCollection(fontCollection))
			th := themeUI.New(theme)

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return tm.LayoutTabs(gtx, th, tabNames)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					switch tm.ActiveTab {
					case 0:
						return tabs.Exams(gtx, th, &selectedTestID, db, tm)
					case 1:
						return tabs.Students(gtx, th, db)
					case 2:
						return uploadCsv.CreateExam(gtx, th, db, tm)
					case 3:
						if selectedTestID != 0 {
							uploadTab.SetTestID(selectedTestID)
						}
						return uploadTab.Layout(gtx, th, db, w)
					case 4:
						return settingTab.Layout(gtx, th, w)
					default:
						return layout.Dimensions{}
					}
				}),
			)
			if tm.ActiveTab == 3 {
				uploadTab.HandleEvent(evt)
			}
			typ.Frame(gtx.Ops)

		case app.DestroyEvent:
			logger.Info("Zatvorenie aplikácie.")
			os.Exit(0)
		}
	}
}
