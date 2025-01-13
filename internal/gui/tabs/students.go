package tabs

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/unit"
)

// Studenti renderuje obsah pre záložku "Študenti".
func Students(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.Label(th, unit.Sp(20), "Toto je záložka Študenti").Layout(gtx)
	})
}
