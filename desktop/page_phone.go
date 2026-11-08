package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

var keypad = []struct{ digit, letters string }{
	{"1", ""}, {"2", "ABC"}, {"3", "DEF"},
	{"4", "GHI"}, {"5", "JKL"}, {"6", "MNO"},
	{"7", "PQRS"}, {"8", "TUV"}, {"9", "WXYZ"},
	{"*", ""}, {"0", "+"}, {"#", ""},
}

// phonePage is the keypad beside the recent calls, as the web UI's phone
// page on the desktop.
func (a *app) phonePage(c *ui.Context, p palette) {
	ui.Row(c).Grow(1).MinHeight(0).Gap(32).Padding(titleBarHeight+4, 28, 28, 28).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(360).Shrink(0).Radius(radiusCard).Background(p.Card).Padding(20, 20, 24).Gap(16).AlignItems(ui.Center).Children(func() {
			a.keypad(c, p)
		})
		ui.Column(c).Grow(1).MinWidth(0).MaxHeightPercent(100).Gap(14).Children(func() {
			a.recents(c, p)
		})
	})
}

// lineChip is a capsule of glass naming a line, whose menu chooses another.
func (a *app) lineChip(c *ui.Context, p palette, lines []Device, chosen string, choose func(id string)) {
	if len(lines) == 0 {
		return
	}
	d := lines[0]
	for _, l := range lines {
		if l.ID == chosen {
			d = l
		}
	}
	chip := ui.Row(c).Height(32).Padding(0, 12).Gap(6).Radius(16).AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Material(control()).Role(ui.RoleMenuButton).Label("线路：" + d.Title())
	chip.Children(func() {
		_, tone := d.Status()
		dot(c, p.tone(tone))
		ui.Text(c, d.Title()).SingleLine()
		if k := lineKind(d); k != "" {
			ui.Text(c, "· "+k).FontSize(12).TextColor(p.Secondary)
		}
		if len(lines) > 1 {
			ui.Icon(c, icChevronDown).FontSize(12).TextColor(p.Secondary)
		}
	})
	if len(lines) > 1 {
		chip.Menu(func(m *ui.Menu) {
			for _, l := range lines {
				if m.Item(l.Title()).Checked(l.ID == d.ID).Chosen() {
					choose(l.ID)
				}
			}
		})
	}
}

func (a *app) keypad(c *ui.Context, p palette) {
	lines := a.readyLines()
	if len(lines) == 0 {
		ui.Row(c).Padding(8, 12).Gap(8).Radius(radiusRow).Background(p.Orange.Alpha(0.14)).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, icWarn).FontSize(15).TextColor(p.Orange)
			ui.Text(c, "没有可通话的线路。在“设备”里开启 WiFi 通话，等 IMS 注册后即可拨打。").FontSize(12).Grow(1)
		})
	} else {
		a.lineChip(c, p, lines, a.dialLine, func(id string) { a.dialLine = id })
	}
	// The number, as large as the iPhone's, shrinking as it grows.
	size := float32(32)
	if len(a.dialed) > 12 {
		size = 24
	}
	ui.Row(c).Height(56).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		ui.Text(c, a.dialed).FontSize(size).SingleLine().Selectable()
	})
	keys := ui.Grid(c).Columns(3).GapX(24).GapY(16).Focusable().Label("拨号键盘")
	keys.Children(func() {
		for _, k := range keypad {
			b := ui.Column(c.Key(k.digit)).Size(78, 78).Radius(39).Center().Background(p.Fill).Cursor(ui.CursorPointer).
				Role(ui.RoleButton).Label(k.digit)
			if b.Pressed() {
				b.Background(p.FillPressed)
			}
			// Keys give a little under the finger and spring back.
			b.AnimatePose(ui.Pose{Scale: map[bool]float32{true: 0.93, false: 1}[b.Pressed()]}, ui.Bouncy)
			b.Children(func() {
				ui.Text(c, k.digit).FontSize(34).LineHeight(1.05)
				if k.letters != "" {
					ui.Text(c, k.letters).FontSize(10).Bold().LetterSpacing(1.6).LineHeight(1.2)
				}
			})
			if b.Clicked() {
				a.dialed += k.digit
			}
		}
	})
	// Typing digits works too, while the keypad has the focus.
	for d := 0; d <= 9; d++ {
		if keys.Shortcut(0, ui.Key0+ui.Key(d)) {
			a.dialed += fmt.Sprint(d)
		}
	}
	if keys.Shortcut(ui.Shift, ui.Key0+ui.Key(8)) {
		a.dialed += "*"
	}
	if keys.Shortcut(ui.Shift, ui.Key0+ui.Key(3)) {
		a.dialed += "#"
	}
	if keys.Shortcut(0, ui.KeyBackspace) && a.dialed != "" {
		r := []rune(a.dialed)
		a.dialed = string(r[:len(r)-1])
	}
	if keys.Shortcut(0, ui.KeyEnter) && a.dialed != "" && !a.callBusy && len(lines) > 0 {
		a.dial(a.dialed)
	}
	ui.Row(c).Gap(24).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(78, 78)
		fill := p.Green
		if a.dialed == "" || len(lines) == 0 {
			fill = p.Green.Alpha(0.4)
		}
		call := roundButton(c, icCall, fill, 78, "呼叫").Disabled(a.dialed == "" || a.callBusy || len(lines) == 0)
		if call.Clicked() {
			a.dial(a.dialed)
		}
		del := ui.Box(c).Size(78, 78).Center().Cursor(ui.CursorPointer).Role(ui.RoleButton).Label("删除")
		if a.dialed != "" {
			del.Children(func() { ui.Icon(c, icDelete).FontSize(26).TextColor(p.Secondary) })
			if del.Clicked() {
				r := []rune(a.dialed)
				a.dialed = string(r[:len(r)-1])
			}
		}
	})
	ui.Text(c, "通话的声音不经过桌面客户端：接通后请在浏览器里的 Halo 通话。").FontSize(11).TextColor(p.Tertiary).TextAlign(ui.Center).MaxWidth(280)
}

func (a *app) recents(c *ui.Context, p palette) {
	missed := 0
	for _, r := range a.calls {
		if r.Missed() {
			missed++
		}
	}
	missedLabel := "未接"
	if missed > 0 {
		missedLabel = fmt.Sprintf("未接 · %d", missed)
	}
	segmented(c, p, &a.phoneTab, 220, "全部", missedLabel).Label("筛选")
	if !a.loadState(c, p, "calls") {
		return
	}
	var list []CallRecord
	for _, r := range a.calls {
		if a.phoneTab == 1 && !r.Missed() {
			continue
		}
		list = append(list, r)
	}
	if len(list) == 0 {
		title := "没有通话记录"
		if a.phoneTab == 1 {
			title = "没有未接来电"
		}
		emptyState(c, p, icPhone, title, "")
		return
	}
	ui.Column(c).Radius(radiusSection).Background(p.Card).Clip().Children(func() {
		for i, r := range list {
			row(c.Key(r.ID), p, i == len(list)-1, func() {
				ic := icOutgoing
				if r.Direction == "incoming" {
					ic = icIncoming
				}
				col := p.Secondary
				if r.Missed() {
					col = p.Red
				}
				ui.Box(c).Width(22).Center().Shrink(0).Children(func() { ui.Icon(c, ic).FontSize(16).TextColor(col) })
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					name := ui.Text(c, displayNumber(r.Number)).Bold().SingleLine()
					if r.Missed() {
						name.TextColor(p.Red)
					}
					ui.Text(c, joinDot(transportName(r.Transport), a.lineName(r.DeviceID), formatDuration(r.DurationSeconds))).FontSize(11).TextColor(p.Secondary).SingleLine()
				})
				ui.Text(c, relTime(r.StartedAt, a.now())).FontSize(12).TextColor(p.Secondary)
				if isDialable(r.Number) {
					b := ui.Box(c).Size(28, 28).Radius(14).Center().Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(fmt.Sprintf("回拨 %s", r.Number)).Tooltip("回拨")
					if b.Hovered() {
						b.Background(p.Fill)
					}
					b.Children(func() { ui.Icon(c, icPhone).FontSize(16).TextColor(c.Theme().Accent) })
					if b.Clicked() {
						a.dialLine = r.DeviceID
						a.dialed = r.Number
						a.dial(r.Number)
					}
				}
			})
		}
	})
}

func (a *app) lineName(id string) string {
	for _, d := range a.devices {
		if d.ID == id {
			return d.Title()
		}
	}
	return ""
}

func transportName(t string) string {
	switch t {
	case "vowifi":
		return "Wi-Fi Calling"
	case "cellular":
		return "蜂窝"
	}
	return ""
}
