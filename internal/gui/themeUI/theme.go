package themeUI

import (
	"image/color"

	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	White       = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	Black       = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	LightGreen  = color.NRGBA{R: 0x8b, G: 0xc3, B: 0x4a, A: 0xff}
	LightRed    = color.NRGBA{R: 0xff, G: 0x73, B: 0x73, A: 0xff}
	LightYellow = color.NRGBA{R: 0xff, G: 0xe0, B: 0x73, A: 0xff}
	LightBlue   = color.NRGBA{R: 0x45, G: 0x89, B: 0xf5, A: 0xff}
	LightPurple = color.NRGBA{R: 0x9c, G: 0x27, B: 0xb0, A: 0xff}
	Red         = color.NRGBA{R: 220, G: 53, B: 69, A: 255}
	Gray        = color.NRGBA{R: 108, G: 117, B: 125, A: 255}
	Green       = color.NRGBA{R: 40, G: 167, B: 69, A: 255}
	LightGray   = color.NRGBA{R: 211, G: 211, B: 211, A: 255}
)

type Theme struct {
	*material.Theme

	LoaderColor           color.NRGBA
	BorderColor           color.NRGBA
	BorderColorFocused    color.NRGBA
	TextColor             color.NRGBA
	ButtonTextColor       color.NRGBA
	SendButtonBgColor     color.NRGBA
	DeleteButtonBgColor   color.NRGBA
	SwitchBgColor         color.NRGBA
	TabInactiveColor      color.NRGBA
	SeparatorColor        color.NRGBA
	SideBarBgColor        color.NRGBA
	SideBarTextColor      color.NRGBA
	TableBorderColor      color.NRGBA
	CheckBoxColor         color.NRGBA
	RequestMethodColor    color.NRGBA
	DropDownMenuBgColor   color.NRGBA
	MenuBgColor           color.NRGBA
	TextSelectionColor    color.NRGBA
	NotificationBgColor   color.NRGBA
	NotificationTextColor color.NRGBA
	ResponseStatusColor   color.NRGBA
	ErrorColor            color.NRGBA
	WarningColor          color.NRGBA
	BadgeBgColor          color.NRGBA
}

func New(material *material.Theme) *Theme {
	t := &Theme{
		Theme:            material,
		SideBarBgColor:   rgb(0x202224),
		SideBarTextColor: rgb(0xffffff),
	}

	t.Theme.TextSize = unit.Sp(14)
	t.LoaderColor = rgb(0x000000)
	t.Theme.Palette.Fg = rgb(0x000000)
	t.Theme.Palette.Bg = rgb(0xffffff)
	t.Theme.Palette.ContrastBg = rgb(0x4589f5)
	t.Theme.Palette.ContrastFg = rgb(0x000000)
	t.BorderColorFocused = rgb(0x4589f5)
	t.BorderColor = rgb(0x6c6f76)
	t.TabInactiveColor = rgb(0x4589f5)
	t.SendButtonBgColor = rgb(0x4589f5)
	t.SwitchBgColor = rgb(0x4589f5)
	t.TextColor = rgb(0x000000)
	t.ButtonTextColor = rgb(0xffffff)
	t.SeparatorColor = rgb(0x9c9c9c)
	t.TableBorderColor = rgb(0xb0b3b8)
	t.CheckBoxColor = rgb(0x4589f5)
	t.RequestMethodColor = rgb(0x007518)
	t.DropDownMenuBgColor = rgb(0x2b2d31)
	t.MenuBgColor = rgb(0x2b2d31)
	t.TextSelectionColor = rgb(0xccd3de)
	t.NotificationBgColor = rgb(0x4589f5)
	t.NotificationTextColor = rgb(0xffffff)
	t.ResponseStatusColor = rgb(0x007518)
	t.ErrorColor = rgb(0xff7373)
	t.WarningColor = rgb(0xffe073)
	t.BadgeBgColor = rgb(0x2b2d31)
	t.DeleteButtonBgColor = rgb(0xff7373)
	return t
}

func (t *Theme) Material() *material.Theme {
	return t.Theme
}

func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}
