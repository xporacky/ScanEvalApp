
package tabmanager

import (
    "gioui.org/layout"
    "gioui.org/widget"
    "gioui.org/widget/material"
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

func (tm *TabManager) LayoutTabs(gtx layout.Context, th *material.Theme, tabNames []string) layout.Dimensions {
    return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
        func() []layout.FlexChild {
            children := make([]layout.FlexChild, len(tm.Buttons))
            for i, btn := range tm.Buttons {
                idx := i
                children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
                    if idx != 3 && btn.Clicked(gtx) {
                        tm.ActiveTab = idx
                    }
                    label := tabNames[idx]
                    if tm.ActiveTab == idx {
                        label += " *"
                    }
                    return material.Button(th, btn, label).Layout(gtx)
                })
            }
            return children
        }()...,
    )
}
