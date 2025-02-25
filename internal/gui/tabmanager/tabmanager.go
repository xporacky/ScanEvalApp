package tabmanager

import (
	"ScanEvalApp/internal/gui/themeUI"
	"ScanEvalApp/internal/gui/widgets"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

type TabManager struct {
	ActiveTab int
	Buttons   []*widget.Clickable
}

func NewTabManager(numTabs int) *TabManager {
	buttons := make([]*widget.Clickable, numTabs)
	for i := range buttons {
		buttons[i] = new(widget.Clickable)
	}
	return &TabManager{ActiveTab: 0, Buttons: buttons}
}

func (tm *TabManager) LayoutTabs(gtx layout.Context, th *themeUI.Theme, tabNames []string) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		func() []layout.FlexChild {
			children := make([]layout.FlexChild, len(tm.Buttons))
			for i, btnTM := range tm.Buttons {
				idx := i
				children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if idx != 3 && btnTM.Clicked(gtx) {
						tm.ActiveTab = idx
					}
					label := tabNames[idx]
					textSize := unit.Sp(14)
					if tm.ActiveTab == idx {
						label += " *"
						textSize = unit.Sp(16)
					}
					btn := widgets.Button(th.Material(), btnTM, nil, widgets.IconPositionEnd, label)
					btn.CornerRadius = unit.Dp(0)
					btn.Background = themeUI.LightGray
					btn.Color = themeUI.Black
					btn.TextSize = textSize
					return btn.Layout(gtx, th)
				})
			}
			return children
		}()...,
	)
}
