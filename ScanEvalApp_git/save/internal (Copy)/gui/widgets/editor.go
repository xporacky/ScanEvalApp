package widgets

import (
	"ScanEvalApp/internal/gui/themeUI"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type EditorField struct {
	Editor      *widget.Editor
	Placeholder string
	BorderColor color.NRGBA
	//Width		unit.Dp
}

func NewEditorField(th *material.Theme, input *widget.Editor, placeholder string) *EditorField {
	e := &EditorField{
		Editor:      input,
		Placeholder: placeholder,
		BorderColor: color.NRGBA{A: 255},
		//Width:       width,
	}
	e.Editor.SingleLine = true
	return e
}

func (e *EditorField) SetText(text string) {
	e.Editor.SetText(text)
}

func (e *EditorField) GetText() string {
	return e.Editor.Text()
}

func (e *EditorField) SetBorderColor(color color.NRGBA) {
	e.BorderColor = color
}

func (e *EditorField) Layout(gtx layout.Context, theme *themeUI.Theme) layout.Dimensions {
	border := widget.Border{
		Color:        e.BorderColor,
		Width:        unit.Dp(2),
		CornerRadius: unit.Dp(4),
	}

	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		//minWidth := gtx.Dp(e.Width)
		//gtx.Constraints.Min.X = minWidth
		return layout.Inset{
			Top:    unit.Dp(4),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Editor(theme.Material(), e.Editor, e.Placeholder).Layout(gtx)
		})
	})
}
