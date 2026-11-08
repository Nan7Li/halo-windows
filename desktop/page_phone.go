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

// phonePage is the keypad beside the recent calls, as Phone on the Mac.
func (a *app) phonePage(c *ui.Context, p palette) {
	ui.Row(c).Grow(1).MinHeight(0).Gap(16).Padding(0, 16, 16, 16).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Width(320).Shrink(0).Radius(16).Background(p.Card).Padding(18).Gap(14).AlignItems(ui.Center).Children(func() {
			a.keypad(c, p)
		})
		ui.Column(c).Grow(1).MinWidth(0).Radius(16).Background(p.Card).Clip().Children(func() {
			a.recents(c, p)
		})
	})
}

func (a *app) keypad(c *ui.Context, p palette) {
	lines := a.readyLines()
	if len(lines) == 0 {
		ui.Row(c).Padding(8, 10).Gap(8).Radius(10).Background(p.Orange.Alpha(0.14)).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, icWarn).FontSize(15).TextColor(p.Orange)
			ui.Text(c, "没有可通话的线路。在“设备”里开启 WiFi 通话，等 IMS 注册后即可拨打。").FontSize(12).Grow(1)
		})
	} else {
		names := make([]string, len(lines))
		chosen := lines[0].Title()
		for i, d := range lines {
			names[i] = d.Title()
			if d.ID == a.dialLine {
				chosen = d.Title()
			}
		}
		if ui.Select(c, &chosen, names).Label("拨号线路").Width(220).Changed() {
			for _, d := range lines {
				if d.Title() == chosen {
					a.dialLine = d.ID
				}
			}
		}
	}
	// The number, as large as the iPhone's, shrinking as it grows.
	size := float32(32)
	if len(a.dialed) > 12 {
		size = 24
	}
	ui.Row(c).Height(52).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		ui.Text(c, a.dialed).FontSize(size).SingleLine().Selectable()
	})
	keys := ui.Grid(c).Columns(3).Gap(14).Focusable().Label("拨号键盘")
	keys.Children(func() {
		for _, k := range keypad {
			b := ui.Column(c.Key(k.digit)).Size(72, 72).Radius(36).Center().Background(p.Fill).Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(k.digit)
			if b.Pressed() {
				b.Background(p.FillPressed)
			}
			b.Children(func() {
				ui.Text(c, k.digit).FontSize(30)
				if k.letters != "" {
					ui.Text(c, k.letters).FontSize(10).Bold().TextColor(p.Secondary).LetterSpacing(1.5)
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
	ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(72, 72)
		call := roundButton(c, icCall, p.Green, 72, "呼叫").Disabled(a.dialed == "" || a.callBusy || len(lines) == 0)
		if call.Clicked() {
			a.dial(a.dialed)
		}
		del := ui.Box(c).Size(72, 72).Center().Cursor(ui.CursorPointer).Role(ui.RoleButton).Label("删除")
		if a.dialed != "" {
			del.Children(func() { ui.Icon(c, icDelete).FontSize(26).TextColor(p.Secondary) })
			if del.Clicked() {
				r := []rune(a.dialed)
				a.dialed = string(r[:len(r)-1])
			}
		}
	})
	ui.Text(c, "通话音频暂不经过桌面客户端：接通后请用浏览器里的 Halo 通话，或使用模组侧的耳机。").FontSize(11).TextColor(p.Tertiary).TextAlign(ui.Center).MaxWidth(260)
}

func (a *app) recents(c *ui.Context, p palette) {
	ui.Row(c).Padding(12, 14).Gap(10).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(p.Separator).Children(func() {
		ui.Text(c, "最近通话").Bold().Grow(1)
		ui.Segmented(c, &a.phoneTab, "全部", "未接").Label("筛选").Width(160)
	})
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
	ui.List(c, nil, len(list), func(i int) {
		r := list[i]
		ui.Row(c.Key(r.ID)).Padding(10, 16).Gap(12).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(p.Separator).Children(func() {
			ic := icOutgoing
			if r.Direction == "incoming" {
				ic = icIncoming
			}
			ui.Icon(c, ic).FontSize(15).TextColor(p.Secondary).Width(16)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				name := ui.Text(c, displayNumber(r.Number)).FontSize(15).Bold().SingleLine()
				if r.Missed() {
					name.TextColor(p.Red)
				}
				ui.Text(c, joinDot(a.lineName(r.DeviceID), transportName(r.Transport), formatDuration(r.DurationSeconds))).FontSize(12).TextColor(p.Secondary).SingleLine()
			})
			ui.Text(c, relTime(r.StartedAt, a.now())).FontSize(12).TextColor(p.Secondary)
			if isDialable(r.Number) {
				b := ui.Box(c).Size(30, 30).Radius(15).Center().Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(fmt.Sprintf("回拨 %s", r.Number)).Tooltip("回拨")
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
	}).Grow(1).Children(func() {
		if len(list) == 0 {
			title := "没有通话记录"
			if a.phoneTab == 1 {
				title = "没有未接来电"
			}
			emptyState(c, p, icPhone, title, "")
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
		return "WiFi 通话"
	case "cellular":
		return "蜂窝"
	}
	return ""
}
