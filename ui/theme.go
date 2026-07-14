package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeMode 是主题模式:跟随系统 / 强制浅色 / 强制深色。
type ThemeMode int

const (
	ModeAuto ThemeMode = iota
	ModeLight
	ModeDark
)

// palette 是一整套配色,深、浅各一份。
type palette struct {
	accent, accentDim, base, card, input, hover, muted color.NRGBA
}

var (
	darkPalette = palette{
		accent:    rgb(0x8B, 0x7C, 0xFF),
		accentDim: rgba(0x8B, 0x7C, 0xFF, 0x55),
		base:      rgb(0x15, 0x16, 0x1E), // 窗口最底层
		card:      rgb(0x20, 0x21, 0x2E), // 卡片/内容背景
		input:     rgb(0x2A, 0x2B, 0x3B), // 按钮/输入框
		hover:     rgb(0x34, 0x35, 0x48),
		muted:     rgb(0x9A, 0x9C, 0xB4), // 次要文字
	}
	lightPalette = palette{
		accent:    rgb(0x6C, 0x5C, 0xE7),
		accentDim: rgba(0x6C, 0x5C, 0xE7, 0x33),
		base:      rgb(0xEC, 0xED, 0xF3),
		card:      rgb(0xFF, 0xFF, 0xFF),
		input:     rgb(0xF1, 0xF2, 0xF8),
		hover:     rgb(0xE2, 0xE3, 0xEF),
		muted:     rgb(0x69, 0x6B, 0x80),
	}
)

func rgb(r, g, b uint8) color.NRGBA     { return color.NRGBA{R: r, G: g, B: b, A: 0xFF} }
func rgba(r, g, b, a uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: a} }

// appTheme 是可在深/浅之间切换的自定义主题,内嵌默认主题作为回落。
type appTheme struct {
	fyne.Theme
	mode ThemeMode
}

// NewAppTheme 创建主题,默认跟随系统。
func NewAppTheme() *appTheme { return &appTheme{Theme: theme.DefaultTheme(), mode: ModeAuto} }

// resolveVariant 把"系统变体 + 当前模式"解析成最终使用的深/浅变体。
func (t *appTheme) resolveVariant(sys fyne.ThemeVariant) fyne.ThemeVariant {
	switch t.mode {
	case ModeLight:
		return theme.VariantLight
	case ModeDark:
		return theme.VariantDark
	default:
		return sys
	}
}

func (t *appTheme) paletteFor(sys fyne.ThemeVariant) palette {
	if t.resolveVariant(sys) == theme.VariantLight {
		return lightPalette
	}
	return darkPalette
}

// currentPalette 依据当前 app 设置(系统变体 + 模式)返回配色,供自定义画布元素取色。
func (t *appTheme) currentPalette() palette {
	sys := theme.VariantDark
	if a := fyne.CurrentApp(); a != nil {
		sys = a.Settings().ThemeVariant()
	}
	return t.paletteFor(sys)
}

func (t *appTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	p := t.paletteFor(v)
	switch n {
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return p.accent
	case theme.ColorNameFocus, theme.ColorNameSelection:
		return p.accentDim
	case theme.ColorNameBackground:
		return p.card
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return p.input
	case theme.ColorNameHover, theme.ColorNamePressed:
		return p.hover
	case theme.ColorNamePlaceHolder:
		return p.muted
	}
	return t.Theme.Color(n, t.resolveVariant(v))
}

func (t *appTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameCardRadius:
		return 12
	case theme.SizeNameButtonRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 6
	}
	return t.Theme.Size(n)
}
