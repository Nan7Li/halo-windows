package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

const (
	titleBarHeight = 52
	sidebarWidth   = 220
	sidebarInset   = 8
)

type pageInfo struct {
	id, title string
	icon      *ui.SVG
}

var pages = []pageInfo{
	{pageToday, "概览", icToday},
	{pageMessages, "信息", icMessages},
	{pagePhone, "电话", icPhone},
	{pageDevices, "设备", icDevices},
	{pageNetwork, "网络", icNetwork},
	{pageLogs, "日志", icLogs},
	{pageSettings, "设置", icSettings},
}

func pageTitle(id string) string {
	for _, p := range pages {
		if p.id == id {
			return p.title
		}
	}
	return ""
}

func (a *app) view(c *ui.Context) {
	p := a.applyTheme(c)
	for _, msg := range a.toast {
		c.Toast(msg)
	}
	a.toast = a.toast[:0]
	switch a.phase {
	case phaseStarting:
		ui.Column(c).Fill().Center().Gap(14).DragWindow().Children(func() {
			ui.Spinner(c)
			ui.Text(c, "正在连接 "+a.prefs.Server).TextColor(p.Secondary)
		})
	case phaseSignedOut:
		a.loginView(c, p)
	default:
		a.mainView(c, p)
	}
}

// wallpaper is the web UI's ambient field: the accent, pink, teal and
// purple glowing from the corners of the grouped background. The glass of
// the sidebar bends and frosts it, as a Mac's sidebar does the desktop. It
// paints on the element behind everything, so cards stay opaque on it.
func wallpaper(c *ui.Context, p palette, e ui.Element) ui.Element {
	k := float32(1)
	if p.dark {
		k = 1.25
	}
	accent := c.Theme().Accent
	glows := []struct {
		x, y, size float32
		col        color
		alpha      float32
	}{
		{0, 0, 0.55, accent, 0.16},
		{1, 0.06, 0.45, p.Pink, 0.11},
		{0.85, 1, 0.55, p.Teal, 0.13},
		{0.08, 0.95, 0.45, p.Purple, 0.10},
	}
	return e.Background(p.Grouped).Draw(func(pt *ui.Painter, r ui.Rect) {
		for _, g := range glows {
			pt.FillRadialGradient(r, ui.RadialGradient{
				Center: ui.Point{X: g.x, Y: g.y}, Radius: g.size * max(r.W, r.H),
				Stops: []ui.GradientStop{{Color: g.col.Alpha(g.alpha * k), At: 0}, {Color: g.col.Alpha(g.alpha * k * 0.45), At: 0.35}, {Color: ui.Transparent, At: 1}},
			}, 0)
		}
	})
}

func (a *app) mainView(c *ui.Context, p palette) {
	wallpaper(c, p, ui.Box(c).Fill()).Children(func() {
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			a.sidebar(c, p)
			// The page scrolls under its toolbar, which floats over it on
			// the scroll edge effect, as macOS 27's.
			ui.Box(c).Grow(1).MinWidth(0).Children(func() {
				ui.Column(c.Key("page:" + a.page)).Fill().Children(func() {
					switch a.page {
					case pageMessages:
						a.messagesPage(c, p)
					case pagePhone:
						a.phonePage(c, p)
					case pageDevices:
						a.devicesPage(c, p)
					case pageNetwork:
						a.networkPage(c, p)
					case pageLogs:
						a.logsPage(c, p)
					case pageSettings:
						a.settingsPage(c, p)
					default:
						a.todayPage(c, p)
					}
				}).Transition(ui.ElementTransition{Enter: &ui.Motion{Y: 10, Opacity: 0}, Animation: &ui.Snappy})
				if a.page != pageMessages {
					if a.page == pageToday || a.page == pageNetwork || a.page == pageSettings {
						scrollEdge(c, titleBarHeight+18, p.Grouped.Alpha(0.55), false)
					}
					a.titleBar(c, p)
				}
			})
		})
		a.callBanner(c, p)
	})
}

// sidebarSel is the identity of the sidebar's selection, which slides
// from the row chosen before to the one chosen now.
type sidebarSel struct{}

// sidebarItem is a row of the sidebar: a tinted icon and a label, on a
// capsule that slides on a spring to the page chosen, as iOS 27's.
func sidebarItem(c *ui.Context, p palette, key any, selected bool, fn func()) ui.Element {
	r := ui.Row(c.Key(key)).Height(30).Padding(0, 9).Gap(9).Radius(radiusItem).AlignItems(ui.Center).
		Cursor(ui.CursorPointer).Role(ui.RoleTab).AccessSelected(selected)
	sel, hover := ui.RGBA(118, 118, 128, 0.18), ui.RGBA(118, 118, 128, 0.08)
	if p.dark {
		sel, hover = ui.RGBA(255, 255, 255, 0.13), ui.RGBA(255, 255, 255, 0.06)
	}
	if !selected && r.Hovered() {
		r.Background(hover)
	}
	return r.Children(func() {
		if selected {
			ui.Box(c.Key(sidebarSel{})).Absolute().Fill().Radius(radiusItem).Background(sel).ZIndex(-1).PassThrough().
				Transition(ui.ElementTransition{Match: sidebarSel{}, Animation: &ui.Snappy, Position: true, Size: true})
		}
		fn()
	})
}

func (a *app) sidebar(c *ui.Context, p palette) {
	t := c.Theme()
	unread := 0
	for _, k := range a.contacts {
		unread += k.UnreadCount
	}
	missed := 0
	for _, r := range a.calls {
		if r.Missed() && a.now().Sub(r.StartedAt) < 24*time.Hour {
			missed++
		}
	}
	pane := ui.Column(c).Width(sidebarWidth).Shrink(0).Margin(sidebarInset, 0, sidebarInset, sidebarInset).
		Radius(radiusSidebar).Material(pane())
	pane.Children(func() {
		// The window controls sit inset over the top of the sidebar on
		// macOS; elsewhere the top is room to drag the window by.
		if top := c.TitleBar().Height - sidebarInset - 4; top > 4 && c.TitleBar().Left > 0 {
			ui.Box(c).Height(top).Shrink(0).DragWindow()
		}
		ui.Row(c).Padding(12, 12, 8, 14).Gap(9).AlignItems(ui.Center).DragWindow().Children(func() {
			haloMark(c, t.Accent, 28)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				ui.Text(c, "Halo").Bold()
				host := a.prefs.Server
				if a.api.Preview() {
					host = "预览模式"
				}
				ui.Text(c, hostOnly(host)).FontSize(11).TextColor(p.Secondary).SingleLine()
			})
		})
		ui.Scroll(c).Grow(1).Padding(2, 8, 8).Gap(1).Label("页面").Children(func() {
			item := func(pg pageInfo, badge int, red bool) {
				sel := a.page == pg.id
				r := sidebarItem(c, p, pg.id, sel, func() {
					ui.Icon(c, pg.icon).FontSize(17).TextColor(t.Accent).Shrink(0)
					ui.Text(c, pg.title).Grow(1).SingleLine()
					if badge > 0 {
						b := ui.Text(c, fmt.Sprint(badge)).FontSize(11).Bold().Padding(1, 7).Radius(9)
						if red {
							b.Background(p.Red).TextColor(ui.RGB(255, 255, 255))
						} else {
							b.TextColor(p.Secondary)
						}
					}
				}).Label(pg.title)
				if r.Clicked() {
					a.goTo(pg.id)
				}
			}
			heading := func(s string) {
				ui.Text(c, s).FontSize(11).Bold().TextColor(p.Secondary).Padding(12, 8, 4)
			}
			calls := missed
			if len(a.live) > 0 {
				calls = len(a.live)
			}
			item(pages[0], 0, false)
			item(pages[2], calls, true)
			item(pages[1], unread, false)
			item(pages[3], 0, false)
			if len(a.devices) > 0 {
				heading("设备")
				for _, d := range a.devices {
					sel := a.page == pageDevices && a.deviceID == d.ID
					_, tone := d.Status()
					r := sidebarItem(c, p, "dev:"+d.ID, sel, func() {
						ui.Box(c).Width(17).Center().Shrink(0).Children(func() {
							fill := p.tone(tone)
							if tone == ToneMuted || tone == ToneDanger {
								fill = p.Tertiary
							}
							dot(c, fill)
						})
						ui.Text(c, d.Title()).Grow(1).SingleLine()
						ui.Text(c, lineKind(d)).FontSize(11).TextColor(p.Tertiary)
					}).Label(d.Title())
					if r.Clicked() {
						a.deviceID = d.ID
						a.goTo(pageDevices)
					}
				}
			}
			heading("网络与系统")
			item(pages[4], 0, false)
			item(pages[5], 0, false)
			item(pages[6], 0, false)
		})
		ui.Row(c).Padding(10, 12, 12, 12).Gap(10).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(p.Separator).Children(func() {
			ui.Avatar(c, a.user, nil).Size(34, 34)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				ui.Text(c, a.user).Bold().SingleLine()
				ui.Text(c, nonEmpty(a.version, "Halo")).FontSize(11).TextColor(p.Secondary).SingleLine()
			})
			b := ui.Box(c).Size(28, 28).Radius(14).Center().Cursor(ui.CursorPointer).Role(ui.RoleButton).Label("退出登录").Tooltip("退出登录")
			if b.Hovered() {
				b.Background(p.Fill)
			}
			b.Children(func() { ui.Icon(c, icSignOut).FontSize(17).TextColor(p.Secondary) })
			if b.Clicked() {
				if a.api.Preview() {
					a.signOut()
				} else {
					a.signOutAsk = true
				}
			}
		})
	})
	if ui.AlertDialog(c, &a.signOutAsk, "退出登录？", "会同时删除保存在钥匙串里的密码。", "取消", "退出登录") == 1 {
		a.signOut()
	}
}

// haloMark is the app's mark: the accent ring around a dot, on a soft disc.
func haloMark(c *ui.Context, accent color, size float32) {
	ui.Box(c).Size(size, size).Radius(size / 2).Center().Shrink(0).Background(accent.Alpha(0.14)).Children(func() {
		ui.Box(c).Size(size*0.62, size*0.62).Radius(size*0.31).Border(size*0.09, accent).Center().Children(func() {
			ui.Box(c).Size(size*0.22, size*0.22).Radius(size * 0.11).Background(accent)
		})
	})
}

// lineKind is how a line carries calls, as the sidebar's devices show it.
func lineKind(d Device) string {
	switch {
	case !d.Running:
		return "离线"
	case d.FlightMode || d.RadioMode == "airplane":
		return "飞行"
	case d.VoWiFiEnabled && d.VoWiFi.IMSReady:
		return "Wi-Fi"
	case d.Healthy:
		return "蜂窝"
	}
	return ""
}

func hostOnly(server string) string {
	for _, pre := range []string{"http://", "https://"} {
		if len(server) > len(pre) && server[:len(pre)] == pre {
			return server[len(pre):]
		}
	}
	return server
}

func (a *app) goTo(page string) {
	if a.page == page {
		return
	}
	a.page = page
	a.prefs.Page = page
	a.store.Save(a.prefs)
	switch page {
	case pageToday:
		a.loadHost()
	case pageMessages:
		a.loadContacts()
	case pagePhone:
		a.loadCalls()
	case pageDevices:
		a.loadDevices()
	case pageNetwork:
		a.loadTunnels()
	case pageLogs:
		a.loadLogs()
	}
}

// titleBar is the page's title, under the window controls of Windows and
// Linux at its right, with the page's refresh button.
func (a *app) titleBar(c *ui.Context, p palette) {
	bar := c.TitleBar()
	ui.Row(c).Absolute().Top(0).Left(0).Right(0).Height(titleBarHeight).AlignItems(ui.Center).Padding(0, bar.Right+14, 0, 28).Gap(10).DragWindow().Children(func() {
		ui.Text(c, pageTitle(a.page)).FontSize(15).Bold().SingleLine()
		ui.Box(c).Grow(1)
		if key := pageDataKey(a.page); key != "" {
			if a.refreshing[key] && a.loaded[key] {
				ui.Spinner(c).Size(14, 14)
			}
			toolbarGroup(c, func() {
				if toolButton(c, p, icRefresh, "刷新").Clicked() {
					a.refreshPage()
				}
			})
		}
	})
}

func pageDataKey(page string) string {
	switch page {
	case pageToday:
		return "host"
	case pageMessages:
		return "contacts"
	case pagePhone:
		return "calls"
	case pageDevices:
		return "devices"
	case pageNetwork:
		return "tunnels"
	case pageLogs:
		return "logs"
	}
	return ""
}

func (a *app) refreshPage() {
	switch a.page {
	case pageToday:
		a.loadHost()
		a.loadDevices()
		a.loadTasks()
		a.loadContacts()
	case pageMessages:
		a.loadContacts()
		if a.selected != "" {
			a.openThread(a.selected, false)
		}
	case pagePhone:
		a.loadCalls()
		a.loadDevices()
	case pageDevices:
		a.loadDevices()
	case pageNetwork:
		a.loadTunnels()
	case pageLogs:
		a.loadLogs()
	}
}

// callBanner floats over the content while a call is ringing or in
// progress, as the call's Live Activity does at the top of an iPhone.
func (a *app) callBanner(c *ui.Context, p palette) {
	if len(a.live) == 0 {
		return
	}
	l := a.live[0]
	line := ""
	for _, d := range a.devices {
		if d.ID == l.DeviceID {
			line = d.Title()
		}
	}
	var state string
	switch l.Phase() {
	case PhaseIncoming:
		state = "来电"
	case PhaseDialing:
		state = "正在呼叫…"
	case PhaseConnecting:
		state = "正在接通…"
	case PhaseActive:
		since := l.StartedAt
		if l.AnsweredAt != nil {
			since = *l.AnsweredAt
		}
		state = formatDuration(int(a.now().Sub(since).Seconds()))
		if state == "" {
			state = "0:00"
		}
		c.After(time.Second)
	}
	left := float32(sidebarWidth + sidebarInset)
	ui.Row(c).Absolute().Top(10).Left(left).Right(0).Justify(ui.Center).PassThrough().Children(func() {
		ui.Row(c).Width(440).Padding(10, 12, 10, 16).Gap(12).AlignItems(ui.Center).Radius(28).
			Material(pane()).Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.18)).Role(ui.RoleStatus).Label("通话").Children(func() {
			avatar(c, p, l.Number, 36)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				ui.Text(c, displayNumber(l.Number)).Bold().SingleLine()
				sub := state
				if line != "" {
					sub += " · " + line
				}
				ui.Text(c, sub).FontSize(12).TextColor(p.Secondary).SingleLine()
			})
			if l.Phase() == PhaseIncoming {
				if roundButton(c, icCall, p.Green, 36, "接听").Disabled(a.callBusy).Clicked() {
					a.answer(l)
				}
			}
			if roundButton(c, icHangup, p.Red, 36, "挂断").Disabled(a.callBusy).Clicked() {
				a.hangup(l)
			}
		})
	})
}

// ---- Shared pieces of pages ----

// pageScroll is the scrolling body of a page.
func pageScroll(c *ui.Context, fn func()) ui.Element {
	return ui.Scroll(c).Grow(1).Padding(titleBarHeight+4, 28, 28, 28).Gap(20).Children(fn)
}

// section is an inset grouped list: a small title over a rounded card.
func section(c *ui.Context, p palette, title string, fn func()) {
	ui.Column(c).Gap(6).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(11).TextColor(p.Secondary).Padding(0, 12)
		}
		ui.Column(c).Radius(radiusSection).Background(p.Card).Clip().Children(fn)
	})
}

// row is a row of a grouped list, with a hairline under it unless last.
func row(c *ui.Context, p palette, last bool, fn func()) ui.Element {
	r := ui.Row(c).MinHeight(40).Padding(6, 14).Gap(10).AlignItems(ui.Center)
	return r.Children(func() {
		fn()
		if !last {
			// The hairline starts after the icon, as iOS's inset rows.
			ui.Box(c).Absolute().Left(48).Right(0).Bottom(0).Height(1).Background(p.Separator)
		}
	})
}

// iconTile is the rounded square of color in Settings' rows.
func iconTile(c *ui.Context, ic *ui.SVG, fill color) {
	ui.Box(c).Size(24, 24).Radius(7).Center().Background(fill).Shrink(0).Children(func() {
		ui.Icon(c, ic).FontSize(14).TextColor(ui.RGB(255, 255, 255))
	})
}

// emptyState is a page or pane with nothing to show.
func emptyState(c *ui.Context, p palette, ic *ui.SVG, title, detail string) {
	ui.Column(c).Grow(1).Center().Gap(8).Padding(40).Children(func() {
		ui.Icon(c, ic).FontSize(40).TextColor(p.Tertiary)
		ui.Text(c, title).FontSize(17).Bold().TextColor(p.Secondary).TextAlign(ui.Center)
		if detail != "" {
			ui.Text(c, detail).FontSize(13).TextColor(p.Secondary).TextAlign(ui.Center).MaxWidth(360)
		}
	})
}

// loadState shows a spinner before the first load of key, or its error.
// It returns true when the page should show its data.
func (a *app) loadState(c *ui.Context, p palette, key string) bool {
	if err := a.loadErr[key]; err != "" && !a.loaded[key] {
		ui.Column(c).Grow(1).Center().Gap(10).Padding(40).Children(func() {
			ui.Icon(c, icWarn).FontSize(36).TextColor(p.Orange)
			ui.Text(c, "无法载入").FontSize(17).Bold()
			ui.Text(c, err).FontSize(13).TextColor(p.Secondary).TextAlign(ui.Center).MaxWidth(420)
			if ui.Button(c, "重试").Clicked() {
				a.refreshPage()
			}
		})
		return false
	}
	if !a.loaded[key] {
		ui.Column(c).Grow(1).Center().Children(func() { ui.Spinner(c) })
		return false
	}
	return true
}

// errorBanner shows a failed refresh over data that loaded before.
func (a *app) errorBanner(c *ui.Context, p palette, key string) {
	if err := a.loadErr[key]; err != "" && a.loaded[key] {
		ui.Row(c).Padding(8, 14).Gap(8).Radius(radiusRow).AlignItems(ui.Center).Background(p.Orange.Alpha(0.14)).Children(func() {
			ui.Icon(c, icWarn).FontSize(15).TextColor(p.Orange)
			ui.Text(c, "刷新失败："+err).FontSize(12).Grow(1)
		})
	}
}

func dot(c *ui.Context, fill color) {
	ui.Box(c).Size(8, 8).Radius(4).Background(fill).Shrink(0)
}
