package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

const (
	titleBarHeight = 52
	sidebarWidth   = 236
	sidebarInset   = 10
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

// wallpaper paints soft light behind the sidebar, which its glass bends
// and frosts, as the desktop shows through a Mac's sidebar.
func wallpaper(c *ui.Context, p palette) {
	t := c.Theme()
	a1, a2 := t.Accent.Alpha(0.22), p.Teal.Alpha(0.16)
	if p.dark {
		a1, a2 = t.Accent.Alpha(0.30), p.Indigo.Alpha(0.26)
	}
	ui.Box(c).Absolute().Fill().PassThrough().RadialGradient(ui.RadialGradient{
		Center: ui.Point{X: 0.05, Y: 0.1}, Radius: 520,
		Stops: []ui.GradientStop{{Color: a1, At: 0}, {Color: ui.Transparent, At: 1}},
	})
	ui.Box(c).Absolute().Fill().PassThrough().RadialGradient(ui.RadialGradient{
		Center: ui.Point{X: 0.0, Y: 0.95}, Radius: 460,
		Stops: []ui.GradientStop{{Color: a2, At: 0}, {Color: ui.Transparent, At: 1}},
	})
}

func (a *app) mainView(c *ui.Context, p palette) {
	ui.Box(c).Fill().Children(func() {
		wallpaper(c, p)
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			a.sidebar(c, p)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				a.titleBar(c, p)
				ui.Column(c).Grow(1).MinHeight(0).Children(func() {
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
				})
			})
		})
		a.callBanner(c, p)
	})
}

func (a *app) sidebar(c *ui.Context, p palette) {
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
		Radius(18).Material(glass.Glass{Clarity: float32(a.prefs.Clarity)})
	pane.Children(func() {
		// The window controls sit inset over the top of the sidebar on
		// macOS; elsewhere this is room to drag the window by.
		ui.Box(c).Height(max(c.TitleBar().Height-sidebarInset, 18)).Shrink(0).DragWindow()
		ui.Row(c).Padding(0, 16, 10).Gap(10).AlignItems(ui.Center).DragWindow().Children(func() {
			ui.Box(c).Size(30, 30).Radius(9).Center().Gradient(c.Theme().Accent, p.Indigo, 135).Children(func() {
				ui.Icon(c, icHalo).FontSize(18).TextColor(ui.RGB(255, 255, 255))
			})
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				ui.Text(c, "Halo").Bold().FontSize(15)
				host := a.prefs.Server
				if a.api.Preview() {
					host = "预览模式"
				}
				ui.Text(c, host).FontSize(11).TextColor(p.Secondary).SingleLine()
			})
		})
		page := a.page
		side := ui.Sidebar(c, &page, func() {
			for _, pg := range pages {
				item := ui.SidebarItem(c, pg.id, pg.icon, pg.title)
				switch {
				case pg.id == pageMessages && unread > 0:
					item.Children(func() { ui.Badge(c, fmt.Sprint(unread)) })
				case pg.id == pagePhone && (missed > 0 || len(a.live) > 0):
					n := missed
					if len(a.live) > 0 {
						n = len(a.live)
					}
					item.Children(func() { ui.Badge(c, fmt.Sprint(n)).Background(p.Red).TextColor(ui.RGB(255, 255, 255)) })
				}
			}
		}).Grow(1).Background(ui.Transparent).Label("页面")
		if side.Changed() {
			a.goTo(page)
		}
		ui.Row(c).Padding(10, 14, 14).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Avatar(c, a.user, nil).Size(28, 28)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() {
				ui.Text(c, a.user).FontSize(13).Bold().SingleLine()
				ui.Text(c, a.version).FontSize(11).TextColor(p.Secondary).SingleLine()
			})
		})
	})
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
	ui.Row(c).Height(titleBarHeight).Shrink(0).AlignItems(ui.Center).Padding(0, bar.Right+16, 0, 22).Gap(10).DragWindow().Children(func() {
		ui.Text(c, pageTitle(a.page)).FontSize(22).Bold().SingleLine()
		ui.Box(c).Grow(1)
		if key := pageDataKey(a.page); key != "" {
			if a.refreshing[key] && a.loaded[key] {
				ui.Spinner(c).Size(14, 14)
			}
			b := glassIconButton(c, icRefresh, "刷新")
			if b.Clicked() {
				a.refreshPage()
			}
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

// glassIconButton is a round button of interactive glass, as the toolbar
// buttons of macOS 26.
func glassIconButton(c *ui.Context, ic *ui.SVG, label string) ui.Element {
	return ui.Box(c).Size(32, 32).Radius(16).Center().Cursor(ui.CursorPointer).
		Material(glass.Glass{Interactive: true}).Role(ui.RoleButton).Label(label).Tooltip(label).Children(func() {
		ui.Icon(c, ic).FontSize(16)
	})
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
			Material(glass.Glass{}).Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.18)).Role(ui.RoleStatus).Label("通话").Children(func() {
			ui.Avatar(c, initials(l.Number), nil).Size(36, 36)
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

// roundButton is a filled circle with a white glyph, as the call buttons.
func roundButton(c *ui.Context, ic *ui.SVG, fill color, size float32, label string) ui.Element {
	b := ui.Box(c).Size(size, size).Radius(size / 2).Center().Background(fill).Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Pressed() {
		b.Background(fill.Mix(ui.RGB(0, 0, 0), 0.18))
	}
	b.Children(func() {
		ui.Icon(c, ic).FontSize(size * 0.48).TextColor(ui.RGB(255, 255, 255))
	})
	return b
}

// ---- Shared pieces of pages ----

// pageScroll is the scrolling body of a page.
func pageScroll(c *ui.Context, fn func()) ui.Element {
	return ui.Scroll(c).Grow(1).Padding(4, 22, 22, 22).Gap(18).Children(fn)
}

// section is an inset grouped list: a small title over a rounded card.
func section(c *ui.Context, p palette, title string, fn func()) {
	ui.Column(c).Gap(6).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(12).TextColor(p.Secondary).Padding(0, 12)
		}
		ui.Column(c).Radius(12).Background(p.Card).Clip().Children(fn)
	})
}

// row is a row of a grouped list, with a hairline under it unless last.
func row(c *ui.Context, p palette, last bool, fn func()) ui.Element {
	r := ui.Row(c).MinHeight(44).Padding(8, 14).Gap(12).AlignItems(ui.Center)
	if !last {
		r.BorderWidth(0, 0, 1, 0).BorderColor(p.Separator)
	}
	return r.Children(fn)
}

// iconTile is the rounded square of color in Settings' rows.
func iconTile(c *ui.Context, ic *ui.SVG, fill color) {
	ui.Box(c).Size(28, 28).Radius(7).Center().Background(fill).Shrink(0).Children(func() {
		ui.Icon(c, ic).FontSize(17).TextColor(ui.RGB(255, 255, 255))
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
		ui.Row(c).Padding(8, 12).Gap(8).Radius(10).AlignItems(ui.Center).Background(p.Orange.Alpha(0.14)).Children(func() {
			ui.Icon(c, icWarn).FontSize(15).TextColor(p.Orange)
			ui.Text(c, "刷新失败："+err).FontSize(12).Grow(1)
		})
	}
}

func dot(c *ui.Context, fill color) {
	ui.Box(c).Size(8, 8).Radius(4).Background(fill).Shrink(0)
}
