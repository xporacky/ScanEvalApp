package widgets

import (
	"ScanEvalApp/internal/gui/themeUI"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// CheckboxWithLabel vykreslí checkbox s labelom vedľa seba a nastaviteľnou veľkosťou fontu
func Checkbox(gtx layout.Context, th *themeUI.Theme, checkbox *widget.Bool, label string, fontSize unit.Sp) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(th.Material(), checkbox, "").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			txt := material.Body1(th.Material(), label)
			txt.TextSize = fontSize
			return txt.Layout(gtx)
		}),
	)
}
