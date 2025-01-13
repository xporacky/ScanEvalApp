package tabs

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/unit"
)

// Upload renderuje obsah pre záložku "Upload CSV".
func Upload(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.Label(th, unit.Sp(20), "Toto je záložka Upload CSV").Layout(gtx)
	})
}
