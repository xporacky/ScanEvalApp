// window.go
package window

import (
    "gioui.org/app"
    "gioui.org/layout"
    "gioui.org/op"
    "gioui.org/widget/material"
    "os"
    "ScanEvalApp/internal/gui/tabmanager"  // Import tabmanager balíka
    "ScanEvalApp/internal/gui/tabs"
    "gorm.io/gorm"
    "ScanEvalApp/internal/logging"
)

func RunWindow(db *gorm.DB) {
    logger := logging.GetLogger()

    w := new(app.Window)
    w.Option(app.Title("ScanEvalApp"))
    var ops op.Ops
    tm := tabmanager.NewTabManager(4) // Vytvor TabManager
    tabNames := []string{"Písomky", "Študenti", "Vytvorenie Písomky", "Upload CSV"}

    for {
        evt := w.Event()
        switch typ := evt.(type) {
        case app.FrameEvent:
            gtx := app.NewContext(&ops, typ)
            ops.Reset()

            th := material.NewTheme()

            layout.Flex{Axis: layout.Vertical}.Layout(gtx,
                layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    return tm.LayoutTabs(gtx, th, tabNames) // Vykreslenie záložiek
                }),
                layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
                    switch tm.ActiveTab {
                    case 0:
                        return tabs.Exams(gtx, th, db)
                    case 1:
                        return tabs.Students(gtx, th, db)
                    case 2:
                        return tabs.CreateTest(gtx, th)
                    case 3:
                        return tabs.Upload(gtx, th,w)
                    default:
                        return layout.Dimensions{}
                    }
                }),
            )
            typ.Frame(gtx.Ops)

        case app.DestroyEvent:
            logger.Info("Zatvorenie aplikácie.")
            os.Exit(0)
        }
    }
}
