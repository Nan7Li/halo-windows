package main

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"
)

// todayPage is the web UI's dashboard on the desktop: four widgets in a
// row, the recent calls beside the recent messages, then the devices.
func (a *app) todayPage(c *ui.Context, p palette) {
	if !a.loadState(c, p, "host") {
		return
	}
	w, _ := c.Size()
	content := w - sidebarWidth - sidebarInset - 56
	pageScroll(c, func() {
		a.errorBanner(c, p, "host")
		ui.Text(c, dateLine(a.now())).FontSize(12).TextColor(p.Secondary).Padding(0, 4).Margin(0, 0, -8)
		widgets := []func(){
			func() { a.linesWidget(c, p) },
			func() { a.tasksWidget(c, p) },
			func() { a.hostWidget(c, p) },
			func() { a.netWidget(c, p) },
		}
		perRow := 4
		if content < 860 {
			perRow = 2
		}
		ui.Column(c).Gap(16).Children(func() {
			for i := 0; i < len(widgets); i += perRow {
				ui.Row(c.Key(i)).Gap(16).AlignItems(ui.Stretch).Children(func() {
					for _, f := range widgets[i:min(i+perRow, len(widgets))] {
						f()
					}
				})
			}
		})
		ui.Row(c).Gap(24).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(8).Children(func() {
				sectionHeader(c, p, "最近通话", func() { a.goTo(pagePhone) })
				a.recentCalls(c, p)
			})
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(8).Children(func() {
				sectionHeader(c, p, "最近信息", func() { a.goTo(pageMessages) })
				a.recentMessages(c, p)
			})
		})
		ui.Column(c).Gap(8).Children(func() {
			sectionHeader(c, p, "设备", nil)
			a.deviceRows(c, p)
		})
	})
}

func dateLine(t interface{ Format(string) string }) string {
	return t.Format("1月2日") + " " + weekday(t.Format("Mon"))
}

func weekday(en string) string {
	return map[string]string{"Mon": "星期一", "Tue": "星期二", "Wed": "星期三", "Thu": "星期四", "Fri": "星期五", "Sat": "星期六", "Sun": "星期日"}[en]
}

// sectionHeader is a section's title, with "全部" taking to its page.
func sectionHeader(c *ui.Context, p palette, title string, all func()) {
	ui.Row(c).Padding(0, 2).AlignItems(ui.End).Children(func() {
		ui.Text(c, title).FontSize(15).Bold().Grow(1)
		if all != nil {
			l := ui.Text(c, "全部").TextColor(c.Theme().Accent).Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(title + " 全部")
			if l.Clicked() {
				all()
			}
		}
	})
}

// widget is a card of the dashboard, as an iOS widget: a tinted title and
// what it shows below it.
func widget(c *ui.Context, p palette, ic *ui.SVG, tint color, title string, fn func()) ui.Element {
	return ui.Column(c).Grow(1).Basis(0).MinWidth(0).Height(168).Padding(16, 18).Gap(8).Radius(radiusCard).Background(p.Card).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, ic).FontSize(15).TextColor(tint)
			ui.Text(c, title).FontSize(12).TextColor(tint)
		})
		fn()
	})
}

func (a *app) linesWidget(c *ui.Context, p palette) {
	online, ims := 0, 0
	for _, d := range a.devices {
		if d.Running && d.Healthy {
			online++
		}
		if d.VoWiFiEnabled && d.VoWiFi.IMSReady {
			ims++
		}
	}
	widget(c, p, icDevices, p.Green, "线路", func() {
		ui.Box(c).Grow(1)
		ui.Row(c).Gap(4).AlignItems(ui.End).Children(func() {
			ui.Textf(c, "%d", online).FontSize(26).Bold()
			ui.Textf(c, "/ %d", len(a.devices)).FontSize(15).Bold().TextColor(p.Tertiary).Padding(0, 0, 3)
		})
		caption := "在线"
		if ims > 0 {
			caption += fmt.Sprintf(" · %d 条 IMS 就绪", ims)
		}
		ui.Text(c, caption).FontSize(11).TextColor(p.Secondary)
		frac := 0.0
		if len(a.devices) > 0 {
			frac = float64(online) / float64(len(a.devices))
		}
		bar(c, p, frac, p.Green)
	})
}

func (a *app) tasksWidget(c *ui.Context, p palette) {
	var next *Task
	for i, t := range a.tasks {
		if t.Enabled && (next == nil || t.NextRunAt.Before(next.NextRunAt)) {
			next = &a.tasks[i]
		}
	}
	widget(c, p, icClock, p.Orange, "即将执行", func() {
		ui.Box(c).Grow(1)
		if next == nil {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icClock).FontSize(14).TextColor(p.Secondary)
				ui.Text(c, "没有计划中的任务").FontSize(12).TextColor(p.Secondary)
			})
		} else {
			ui.Text(c, next.Name).FontSize(15).Bold().MaxLines(2)
			ui.Text(c, untilTime(next.NextRunAt, a.now())).FontSize(12).TextColor(p.Secondary)
		}
		if !a.api.Preview() {
			ui.Link(c, "管理任务", a.prefs.Server+"/automatic-tasks")
		}
	})
}

func (a *app) hostWidget(c *ui.Context, p palette) {
	h := a.host
	widget(c, p, icServer, p.Blue, "主机", func() {
		ui.Row(c).Grow(1).Gap(14).AlignItems(ui.Center).Children(func() {
			rings := []struct {
				label string
				v     float64
				col   color
			}{
				{"CPU", h.Perf.CPUPercent, p.Pink},
				{"内存", h.Perf.MemoryPercent, p.Green},
				{"磁盘", h.Perf.DiskPercent, p.Teal},
			}
			ui.Box(c).Size(84, 84).Shrink(0).Label(fmt.Sprintf("CPU %.0f%%，内存 %.0f%%，磁盘 %.0f%%", h.Perf.CPUPercent, h.Perf.MemoryPercent, h.Perf.DiskPercent)).
				Draw(func(pt *ui.Painter, r ui.Rect) {
					cx, cy := r.X+r.W/2, r.Y+r.H/2
					for i, ring := range rings {
						rad := r.W/2 - 6 - float32(i)*12
						pt.StrokePath(arc(cx, cy, rad, 1), 9, ring.col.Alpha(0.22))
						if f := float32(ring.v / 100); f > 0.005 {
							pt.StrokePath(arc(cx, cy, rad, min(f, 1)), 9, ring.col)
						}
					}
				})
			ui.Column(c).Gap(4).MinWidth(0).Children(func() {
				for _, ring := range rings {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						dot(c, ring.col)
						ui.Text(c, ring.label).FontSize(11).TextColor(p.Secondary)
						ui.Textf(c, "%.0f%%", ring.v).FontSize(11).Bold()
					})
				}
			})
		})
		ui.Text(c, joinDot(h.Host.BoardModel, h.Host.CPUModel)).FontSize(11).TextColor(p.Tertiary).SingleLine()
	})
}

// arc is a circle's outline from twelve o'clock, clockwise, frac of it.
func arc(cx, cy, r float32, frac float32) *ui.Path {
	const k = 0.5522847498
	path := &ui.Path{}
	path.MoveTo(cx, cy-r)
	path.CubeTo(cx+r*k, cy-r, cx+r, cy-r*k, cx+r, cy)
	path.CubeTo(cx+r, cy+r*k, cx+r*k, cy+r, cx, cy+r)
	path.CubeTo(cx-r*k, cy+r, cx-r, cy+r*k, cx-r, cy)
	path.CubeTo(cx-r, cy-r*k, cx-r*k, cy-r, cx, cy-r)
	if frac < 1 {
		path.Trim(0, frac)
	}
	return path
}

func (a *app) netWidget(c *ui.Context, p palette) {
	h := a.host
	widget(c, p, icNetwork, p.Teal, "网络", func() {
		rate := func(ic *ui.SVG, label string, v float64, col color) {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(22, 22).Radius(11).Center().Background(col.Alpha(0.15)).Children(func() {
					ui.Icon(c, ic).FontSize(13).TextColor(col)
				})
				ui.Text(c, label).FontSize(12).TextColor(p.Secondary).Grow(1)
				ui.Text(c, formatRate(v)).Bold()
			})
		}
		rate(icDown, "下载", h.Perf.NetRxBps, p.Teal)
		rate(icUp, "上传", h.Perf.NetTxBps, p.Blue)
		hist := a.netHist
		ui.Box(c).Grow(1).MinHeight(24).Label("流量曲线").Draw(func(pt *ui.Painter, r ui.Rect) {
			if len(hist) < 2 {
				return
			}
			peak := 1.0
			for _, s := range hist {
				peak = math.Max(peak, math.Max(s[0], s[1]))
			}
			line := func(idx int) *ui.Path {
				path := &ui.Path{}
				for i, s := range hist {
					x := r.X + r.W*float32(i)/float32(len(hist)-1)
					y := r.Y + r.H - r.H*0.9*float32(s[idx]/peak) - 1
					if i == 0 {
						path.MoveTo(x, y)
					} else {
						path.LineTo(x, y)
					}
				}
				return path
			}
			area := line(0)
			area.LineTo(r.X+r.W, r.Y+r.H).LineTo(r.X, r.Y+r.H).Close()
			pt.FillPathGradient(area, ui.LinearGradient{From: p.Teal.Alpha(0.35), To: p.Teal.Alpha(0.04), Angle: 180})
			pt.StrokePath(line(0), 1.5, p.Teal)
			pt.StrokePath(line(1).Dash(0, 4, 3), 1.2, p.Blue)
		})
	})
}

func (a *app) recentCalls(c *ui.Context, p palette) {
	ui.Column(c).Radius(radiusSection).Background(p.Card).Clip().Children(func() {
		if len(a.calls) == 0 {
			ui.Text(c, "没有通话记录").TextColor(p.Secondary).Padding(14)
			return
		}
		n := min(len(a.calls), 4)
		for i, r := range a.calls[:n] {
			row(c.Key(r.ID), p, i == n-1, func() {
				fill, ic := p.Green, icPhone
				if r.Missed() {
					fill, ic = p.Red, icIncoming
				}
				iconTile(c, ic, fill)
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					name := ui.Text(c, displayNumber(r.Number)).Bold().SingleLine()
					if r.Missed() {
						name.TextColor(p.Red)
					}
					ui.Text(c, joinDot(a.lineName(r.DeviceID), transportName(r.Transport))).FontSize(11).TextColor(p.Secondary).SingleLine()
				})
				ui.Text(c, relTime(r.StartedAt, a.now())).FontSize(12).TextColor(p.Secondary)
			})
		}
	})
}

func (a *app) recentMessages(c *ui.Context, p palette) {
	ui.Column(c).Radius(radiusSection).Background(p.Card).Clip().Children(func() {
		if len(a.contacts) == 0 {
			ui.Text(c, "没有短信").TextColor(p.Secondary).Padding(14)
			return
		}
		n := min(len(a.contacts), 4)
		for i, k := range a.contacts[:n] {
			r := row(c.Key(k.Key()), p, i == n-1, func() {
				ui.Box(c).Absolute().Left(4).Top(0).Bottom(0).Width(6).Center().Children(func() {
					if k.UnreadCount > 0 {
						ui.Box(c).Size(7, 7).Radius(3.5).Background(c.Theme().Accent)
					}
				})
				avatar(c, p, k.Title(), 24)
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					ui.Text(c, k.Title()).Bold().SingleLine()
					ui.Text(c, k.LastMessage).FontSize(11).TextColor(p.Secondary).SingleLine()
				})
				ui.Text(c, relTime(k.LastTime, a.now())).FontSize(12).TextColor(p.Secondary)
			}).Cursor(ui.CursorPointer)
			if r.Clicked() {
				a.goTo(pageMessages)
				a.openThread(k.Key(), true)
			}
		}
	})
}

func (a *app) deviceRows(c *ui.Context, p palette) {
	ui.Column(c).Radius(radiusSection).Background(p.Card).Clip().Children(func() {
		if len(a.devices) == 0 {
			ui.Text(c, "还没有设备").TextColor(p.Secondary).Padding(14)
			return
		}
		for i, d := range a.devices {
			r := row(c.Key(d.ID), p, i == len(a.devices)-1, func() {
				_, tone := d.Status()
				fill := p.Green
				if tone != ToneGood {
					fill = p.Gray
				}
				iconTile(c, icDevices, fill)
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					ui.Text(c, d.Title()).SingleLine()
					sub := joinDot(d.Modem.Operator, d.Modem.NetworkMode, d.PublicIP)
					if !d.Running {
						sub = "离线"
					} else if d.FlightMode {
						sub = joinDot(d.Modem.Operator, "飞行模式")
					}
					ui.Text(c, sub).FontSize(11).TextColor(p.Secondary).SingleLine()
				})
				if d.VoWiFiEnabled && d.VoWiFi.IMSReady {
					pill(c, "Wi-Fi Calling", p.Green)
				}
				signalBars(c, p, d.Bars())
				ui.Icon(c, icChevron).FontSize(13).TextColor(p.Tertiary)
			}).Cursor(ui.CursorPointer)
			if r.Clicked() {
				a.deviceID = d.ID
				a.goTo(pageDevices)
			}
		}
	})
}

// pill is a short status in a capsule of its color.
func pill(c *ui.Context, s string, col color) {
	ui.Text(c, s).FontSize(11).TextColor(col).Background(col.Alpha(0.14)).Padding(2, 8).Radius(9).SingleLine()
}

// avatar is a contact's picture: initials on gray, or a person for numbers.
func avatar(c *ui.Context, p palette, name string, size float32) {
	ini := initials(name)
	ui.Box(c).Size(size, size).Radius(size / 2).Center().Shrink(0).
		LinearGradient(ui.LinearGradient{Angle: 180, From: ui.Hex("#A5A9B1"), To: ui.Hex("#868A92")}).Children(func() {
		if ini == "#" {
			ui.Icon(c, icPerson).FontSize(size * 0.62).TextColor(ui.RGB(255, 255, 255))
			return
		}
		ui.Text(c, ini).FontSize(size * 0.42).Bold().TextColor(ui.RGB(255, 255, 255))
	})
}

// bar is a capsule filled to frac.
func bar(c *ui.Context, p palette, frac float64, fill color) {
	frac = min(max(frac, 0), 1)
	ui.Box(c).Height(6).Radius(3).Background(p.Fill).Clip().Children(func() {
		ui.Box(c).Absolute().Top(0).Bottom(0).Left(0).WidthPercent(float32(frac * 100)).Radius(3).Background(fill)
	})
}

// signalBars draws 0–4 bars, as the status bar's cellular signal.
func signalBars(c *ui.Context, p palette, n int) {
	ui.Row(c).Gap(2).AlignItems(ui.End).Height(12).Shrink(0).Label(fmt.Sprintf("信号 %d 格", n)).Children(func() {
		for i := range 4 {
			fill := p.Label
			if i >= n {
				fill = p.Tertiary.Alpha(0.5)
			}
			ui.Box(c).Size(3, 4+float32(i)*2.7).Radius(1).Background(fill)
		}
	})
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func joinDot(parts ...string) string {
	out := ""
	for _, s := range parts {
		if s == "" || s == "—" {
			continue
		}
		if out != "" {
			out += " · "
		}
		out += s
	}
	return out
}
