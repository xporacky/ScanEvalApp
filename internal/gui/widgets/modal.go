package widgets

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	//"gioui.org/widget/material"
	//"gioui.org/x/component"
	"ScanEvalApp/internal/gui/themeUI"
	//"fmt"
	//"strings"
	"gioui.org/op/paint"
	"image/color"
	"image"
	"gioui.org/op/clip"

)

type Modal struct {
	Visible    	bool
	CloseButton widget.Clickable
	Content     layout.Widget
	SetCloseBtnEnable bool
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
			bgColor := color.NRGBA{R: 255, G: 255, B: 255, A: 240} // Takmer nepriehľadné biele pozadie

			size := gtx.Constraints.Max
			rect := image.Rect(0, 0, size.X, size.Y)
			paint.FillShape(gtx.Ops, bgColor, clip.Rect(rect).Op())
			return layout.Dimensions{Size: size}
		}),

		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(50), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Main content of the modal
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if m.Content != nil {
							return m.Content(gtx) // vložený obsah
						}
						return layout.Dimensions{}
					}),
					
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Close button
						return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							//btn := widgets.Button(theme.Material(), &m.CloseButton, "Zavrieť")
							btn := Button(theme.Theme, &m.CloseButton, CloseIcon, IconPositionStart, "Zavrieť")
							if(m.SetCloseBtnEnable){
								btn.Background = themeUI.Gray
								btn.Color = themeUI.White
							}else{
								btn.Background = themeUI.LightGray
								btn.Color = themeUI.White
							}
							
							if(m.SetCloseBtnEnable){
								if (m.CloseButton.Clicked(gtx)) {
									m.Visible = false // Close the modal
								}}
							return btn.Layout(gtx, theme)
						})
					}),
				)
			})
		}),
	)
}


func (m *Modal) Layout(gtx layout.Context, theme *themeUI.Theme) layout.Dimensions {
	ops := op.Record(gtx.Ops)
	dims := m.layout(gtx, theme)
	defer op.Defer(gtx.Ops, ops.Stop())

	return dims
}
