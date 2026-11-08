package main

import "github.com/egoist/mygo/ui"

// palette holds the colors of Apple's system palette that the views use
// beyond the theme's: grouped backgrounds, fills and the system colors.
type palette struct {
	dark bool
	// Grouped is the window behind content; Card the inset grouped rows
	// on it; Fill a control's fill (keypad keys, search fields).
	Grouped, Card, Fill, FillPressed, Separator  color
	Label, Secondary, Tertiary                   color
	Blue, Green, Orange, Red, Indigo, Teal, Gray color
	Pink, Purple, Yellow                         color
	BubbleOther                                  color
}

type color = ui.Color

func paletteFor(dark bool) palette {
	if dark {
		return palette{
			dark:    true,
			Grouped: ui.Hex("#000000"), Card: ui.Hex("#1C1C1E"),
			Fill: ui.RGBA(118, 118, 128, 0.24), FillPressed: ui.RGBA(118, 118, 128, 0.42),
			Separator: ui.RGBA(84, 84, 88, 0.6),
			Label:     ui.Hex("#FFFFFF"), Secondary: ui.RGBA(235, 235, 245, 0.6), Tertiary: ui.RGBA(235, 235, 245, 0.3),
			Blue: ui.Hex("#0A84FF"), Green: ui.Hex("#30D158"), Orange: ui.Hex("#FF9F0A"), Red: ui.Hex("#FF453A"),
			Indigo: ui.Hex("#5E5CE6"), Teal: ui.Hex("#40C8E0"), Gray: ui.Hex("#8E8E93"),
			Pink: ui.Hex("#FF375F"), Purple: ui.Hex("#BF5AF2"), Yellow: ui.Hex("#FFD60A"),
			BubbleOther: ui.Hex("#26252A"),
		}
	}
	return palette{
		Grouped: ui.Hex("#F2F2F7"), Card: ui.Hex("#FFFFFF"),
		Fill: ui.RGBA(118, 118, 128, 0.12), FillPressed: ui.RGBA(118, 118, 128, 0.26),
		Separator: ui.RGBA(60, 60, 67, 0.18),
		Label:     ui.Hex("#000000"), Secondary: ui.RGBA(60, 60, 67, 0.6), Tertiary: ui.RGBA(60, 60, 67, 0.3),
		Blue: ui.Hex("#007AFF"), Green: ui.Hex("#34C759"), Orange: ui.Hex("#FF9500"), Red: ui.Hex("#FF3B30"),
		Indigo: ui.Hex("#5856D6"), Teal: ui.Hex("#30B0C7"), Gray: ui.Hex("#8E8E93"),
		Pink: ui.Hex("#FF2D55"), Purple: ui.Hex("#AF52DE"), Yellow: ui.Hex("#FFCC00"),
		BubbleOther: ui.Hex("#E9E9EB"),
	}
}

func (p palette) tone(t Tone) color {
	switch t {
	case ToneGood:
		return p.Green
	case ToneWarning:
		return p.Orange
	case ToneDanger:
		return p.Red
	}
	return p.Secondary
}

// accents are the accent colors offered in Settings, as iOS's, in light and
// dark. The empty id follows the system's accent.
var accents = []struct {
	id, name    string
	light, dark color
}{
	{"", "跟随系统", ui.Color{}, ui.Color{}},
	{"blue", "蓝色", ui.Hex("#007AFF"), ui.Hex("#0A84FF")},
	{"indigo", "靛蓝", ui.Hex("#5856D6"), ui.Hex("#5E5CE6")},
	{"purple", "紫色", ui.Hex("#AF52DE"), ui.Hex("#BF5AF2")},
	{"pink", "粉色", ui.Hex("#FF2D55"), ui.Hex("#FF375F")},
	{"orange", "橙色", ui.Hex("#FF9500"), ui.Hex("#FF9F0A")},
	{"green", "绿色", ui.Hex("#34C759"), ui.Hex("#30D158")},
	{"teal", "青色", ui.Hex("#30B0C7"), ui.Hex("#40C8E0")},
	{"graphite", "石墨", ui.Hex("#8E8E93"), ui.Hex("#98989D")},
}

// applyTheme sets the frame's theme from the system's and the user's
// appearance and accent, and returns the palette that goes with it.
func (a *app) applyTheme(c *ui.Context) palette {
	sys := *c.Theme()
	dark := sys.Dark
	switch a.prefs.Appearance {
	case "light":
		dark = false
	case "dark":
		dark = true
	}
	var t ui.Theme
	if dark == sys.Dark {
		t = sys
	} else if dark {
		t = *ui.DarkTheme()
		t.Accent, t.AccentHover, t.AccentPressed, t.AccentText = sys.Accent, sys.AccentHover, sys.AccentPressed, sys.AccentText
	} else {
		t = *ui.LightTheme()
		t.Accent, t.AccentHover, t.AccentPressed, t.AccentText = sys.Accent, sys.AccentHover, sys.AccentPressed, sys.AccentText
	}
	p := paletteFor(dark)
	for _, ac := range accents {
		if ac.id != "" && ac.id == a.prefs.Accent {
			acc := ac.light
			if dark {
				acc = ac.dark
			}
			t.Accent = acc
			t.AccentHover = acc.Mix(ui.RGB(255, 255, 255), 0.12)
			t.AccentPressed = acc.Mix(ui.RGB(0, 0, 0), 0.15)
			t.AccentText = ui.RGB(255, 255, 255)
		}
	}
	t.Background = p.Grouped
	t.Surface = p.Card
	t.Text = p.Label
	t.TextMuted = p.Secondary
	t.Border = p.Separator
	t.Danger = p.Red
	t.Success = p.Green
	t.Warning = p.Orange
	t.Radius = 7
	// macOS's body text (SF Pro 13): the web UI's desktop metrics.
	t.FontSize = 13
	t.InactiveAccent = t.Accent
	c.SetTheme(&t)
	return p
}
