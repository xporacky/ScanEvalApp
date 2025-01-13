package tabs

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/unit"
)

// Vytvorenie renderuje obsah pre záložku "Vytvorenie Písomky".
func CreateTest(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.Label(th, unit.Sp(20), "Toto je záložka Vytvorenie Písomky").Layout(gtx)
	})
}
