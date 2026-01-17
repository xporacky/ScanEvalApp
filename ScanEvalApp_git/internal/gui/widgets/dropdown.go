// In widgets/dropdown.go
package widgets

import (
"ScanEvalApp/internal/gui/themeUI"
"fmt"

"gioui.org/layout"
"gioui.org/unit"
"gioui.org/widget"
"gioui.org/widget/material"
)

type DropdownStyle struct {
Theme      *themeUI.Theme
Options    []string
Selected   *int
Label      string
dropdown   widget.Enum
}

func Dropdown(th *themeUI.Theme, options []string, selected *int, label string) DropdownStyle {
return DropdownStyle{
Theme:    th,
Options:  options,
Selected: selected,
Label:    label,
}
}

func (d DropdownStyle) Layout(gtx layout.Context) layout.Dimensions {
return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// Label
return material.Label(d.Theme.Material(), unit.Sp(14), d.Label).Layout(gtx)
}),
layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// Dropdown options
var children []layout.FlexChild

for i, option := range d.Options {
i, option := i, option // Capture loop variables
children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
// Create a radio button for each option
btn := material.RadioButton(d.Theme.Material(), &d.dropdown, option, fmt.Sprintf("%s", option))
if d.Selected != nil && i == *d.Selected {
btn = material.RadioButton(d.Theme.Material(), &d.dropdown, option, fmt.Sprintf("%s", option))
// This needs to be synchronized with the actual selection
}
return btn.Layout(gtx)
}))
}

return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}),
)
}

// Alternative: Simple selector using buttons
type OptionSelectorStyle struct {
Theme     *themeUI.Theme
Options   []string
Selected  *int
Label     string
buttons   []widget.Clickable
}

func OptionSelector(th *themeUI.Theme, options []string, selected *int, label string) OptionSelectorStyle {
return OptionSelectorStyle{
Theme:    th,
Options:  options,
Selected: selected,
Label:    label,
buttons:  make([]widget.Clickable, len(options)),
}
}

func (os OptionSelectorStyle) Layout(gtx layout.Context) layout.Dimensions {
return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
layout.Rigid(func(gtx layout.Context) layout.Dimensions {
return material.Label(os.Theme.Material(), unit.Sp(14), os.Label).Layout(gtx)
}),
layout.Rigid(func(gtx layout.Context) layout.Dimensions {
var children []layout.FlexChild

for i, option := range os.Options {
i, option := i, option // Capture loop variables
children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
btn := material.Button(os.Theme.Material(), &os.buttons[i], option)

// Change style based on selection
if os.Selected != nil && i == *os.Selected {
btn.Background = themeUI.LightBlue
btn.Color = themeUI.White
} else {
btn.Background = themeUI.LightGray
btn.Color = themeUI.Black
}

if os.buttons[i].Clicked(gtx) {
*os.Selected = i
}

return btn.Layout(gtx)
}))
}

return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceAround}.Layout(gtx, children...)
}),
)
}
