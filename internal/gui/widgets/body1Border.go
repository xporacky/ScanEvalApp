package widgets

import (
	"image/color"

	"ScanEvalApp/internal/gui/themeUI"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func Body1Border(gtx layout.Context, th *themeUI.Theme, txt string) layout.Dimensions {
	border := widget.Border{
		Color: color.NRGBA{A: 255},
		Width: unit.Dp(2),
	}
	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Body1(th.Material(), txt).Layout(gtx)
		})
	})
}
