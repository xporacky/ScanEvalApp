package widgets

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	//"gioui.org/x/component"
	"ScanEvalApp/internal/gui/themeUI"

)

type Modal struct {
	Visible    bool
	CloseButton widget.Clickable
}

func NewModal() *Modal {
	return &Modal{}
}

func (m *Modal) layout(gtx layout.Context, theme *themeUI.Theme) layout.Dimensions {
	if !m.Visible {
		return layout.Dimensions{}
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// Main content of the modal
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Title
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return material.Label(theme.Material(), unit.Sp(20), "Toto je modál").Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Close button
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(theme.Material(), &m.CloseButton, "Zavrieť")
						if m.CloseButton.Clicked(gtx) {
							m.Visible = false // Close the modal
						}
						return btn.Layout(gtx)
					})
				}),
			)
		}),
	)
}

func (m *Modal) Layout(gtx layout.Context, theme *themeUI.Theme) layout.Dimensions {
	ops := op.Record(gtx.Ops)
	dims := m.layout(gtx, theme)
	defer op.Defer(gtx.Ops, ops.Stop())

	return dims
}
