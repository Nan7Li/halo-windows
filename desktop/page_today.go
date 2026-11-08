package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// todayPage is the dashboard: widgets in two columns, as iOS's widgets.
func (a *app) todayPage(c *ui.Context, p palette) {
	if !a.loadState(c, p, "host") {
		return
	}
	pageScroll(c, func() {
		a.errorBanner(c, p, "host")
		w, _ := c.Size()
		cols := 2
		if w < 700 {
			cols = 1
		}
		ui.Row(c).Gap(16).AlignItems(ui.Start).Children(func() {
			left := []func(){func() { a.hostWidget(c, p) }, func() { a.messagesWidget(c, p) }}
			right := []func(){func() { a.linesWidget(c, p) }, func() { a.callsWidget(c, p) }, func() { a.tasksWidget(c, p) }}
			if cols == 1 {
				ui.Column(c).Grow(1).Gap(16).Children(func() {
					for _, f := range append(left, right...) {
						f()
					}
				})
				return
			}
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				for _, f := range left {
					f()
				}
			})
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				for _, f := range right {
					f()
				}
			})
		})
	})
}

// widget is a rounded card with a title, as an iOS widget.
func widget(c *ui.Context, p palette, ic *ui.SVG, tint color, title string, fn func()) ui.Element {
	return ui.Column(c).Padding(14, 16, 16).Gap(10).Radius(18).Background(p.Card).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, ic).FontSize(15).TextColor(tint)
			ui.Text(c, title).FontSize(13).Bold().TextColor(tint)
		})
		fn()
	})
}

func (a *app) hostWidget(c *ui.Context, p palette) {
	h := a.host
	widget(c, p, icServer, p.Teal, "主机", func() {
		ui.Text(c, nonEmpty(h.Host.BoardModel, "Halo 主机")).FontSize(17).Bold()
		ui.Text(c, h.Host.CPUModel).FontSize(12).TextColor(p.Secondary)
		gauge := func(ic *ui.SVG, label string, pct float64, detail string, fill color) {
			ui.Column(c).Gap(5).Children(func() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, ic).FontSize(14).TextColor(p.Secondary)
					ui.Text(c, label).FontSize(13).Grow(1)
					ui.Text(c, detail).FontSize(12).TextColor(p.Secondary)
					ui.Textf(c, "%.0f%%", pct).FontSize(13).Bold().Width(40).TextAlign(ui.End)
				})
				bar(c, p, pct/100, fill)
			})
		}
		mem := ""
		if h.Perf.MemoryTotalBytes > 0 {
			mem = formatBytes(float64(h.Perf.MemoryUsedBytes)) + " / " + formatBytes(float64(h.Perf.MemoryTotalBytes))
		}
		disk := ""
		if h.Perf.DiskTotalBytes > 0 {
			disk = formatBytes(float64(h.Perf.DiskUsedBytes)) + " / " + formatBytes(float64(h.Perf.DiskTotalBytes))
		}
		gauge(icCPU, "处理器", h.Perf.CPUPercent, "", levelColor(p, h.Perf.CPUPercent, c.Theme().Accent))
		gauge(icMemory, "内存", h.Perf.MemoryPercent, mem, levelColor(p, h.Perf.MemoryPercent, p.Green))
		gauge(icDisk, "存储", h.Perf.DiskPercent, disk, levelColor(p, h.Perf.DiskPercent, p.Indigo))
		ui.Row(c).Gap(16).Padding(4, 0, 0).Children(func() {
			for _, x := range []struct {
				ic    *ui.SVG
				label string
				v     float64
			}{{icDown, "下行", h.Perf.NetRxBps}, {icUp, "上行", h.Perf.NetTxBps}} {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Grow(1).Basis(0).Children(func() {
					ui.Icon(c, x.ic).FontSize(14).TextColor(p.Secondary)
					ui.Text(c, x.label).FontSize(12).TextColor(p.Secondary)
					ui.Text(c, formatRate(x.v)).FontSize(13).Bold()
				})
			}
		})
	})
}

func levelColor(p palette, pct float64, normal color) color {
	switch {
	case pct >= 90:
		return p.Red
	case pct >= 75:
		return p.Orange
	}
	return normal
}

// bar is a capsule filled to frac.
func bar(c *ui.Context, p palette, frac float64, fill color) {
	frac = min(max(frac, 0), 1)
	ui.Box(c).Height(6).Radius(3).Background(p.Fill).Clip().Children(func() {
		ui.Box(c).Absolute().Top(0).Bottom(0).Left(0).WidthPercent(float32(frac * 100)).Radius(3).Background(fill)
	})
}

func (a *app) linesWidget(c *ui.Context, p palette) {
	widget(c, p, icDevices, p.Green, "线路", func() {
		if len(a.devices) == 0 {
			ui.Text(c, "还没有设备").TextColor(p.Secondary).FontSize(13)
			return
		}
		for i, d := range a.devices {
			ui.Row(c.Key(d.ID)).Gap(10).AlignItems(ui.Center).Padding(4, 0).Cursor(ui.CursorPointer).Children(func() {
				status, tone := d.Status()
				dot(c, p.tone(tone))
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					ui.Text(c, d.Title()).FontSize(14).Bold().SingleLine()
					ui.Text(c, joinDot(d.Modem.Operator, d.Modem.NetworkMode, status)).FontSize(12).TextColor(p.Secondary).SingleLine()
				})
				signalBars(c, p, d.Bars())
			}).OnClick(func() {
				a.deviceID = d.ID
				a.goTo(pageDevices)
			})
			if i < len(a.devices)-1 {
				ui.Box(c).Height(1).Background(p.Separator)
			}
		}
	})
}

// signalBars draws 0–4 bars, as the status bar's cellular signal.
func signalBars(c *ui.Context, p palette, n int) {
	ui.Row(c).Gap(2).AlignItems(ui.End).Height(14).Shrink(0).Label(fmt.Sprintf("信号 %d 格", n)).Children(func() {
		for i := range 4 {
			fill := p.Label
			if i >= n {
				fill = p.Fill
			}
			ui.Box(c).Size(4, float32(5+i*3)).Radius(1.5).Background(fill)
		}
	})
}

func (a *app) messagesWidget(c *ui.Context, p palette) {
	widget(c, p, icMessages, p.Green, "信息", func() {
		if len(a.contacts) == 0 {
			ui.Text(c, "没有短信").TextColor(p.Secondary).FontSize(13)
			return
		}
		n := min(len(a.contacts), 4)
		for i, k := range a.contacts[:n] {
			ui.Row(c.Key(k.Key())).Gap(10).AlignItems(ui.Center).Padding(4, 0).Cursor(ui.CursorPointer).Children(func() {
				ui.Avatar(c, initials(k.Title()), nil).Size(34, 34)
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					ui.Row(c).Gap(6).Children(func() {
						ui.Text(c, k.Title()).FontSize(14).Bold().SingleLine().Grow(1)
						ui.Text(c, relTime(k.LastTime, a.now())).FontSize(12).TextColor(p.Secondary)
					})
					ui.Text(c, k.LastMessage).FontSize(12).TextColor(p.Secondary).SingleLine()
				})
				if k.UnreadCount > 0 {
					dot(c, c.Theme().Accent)
				}
			}).OnClick(func() {
				a.goTo(pageMessages)
				a.openThread(k.Key(), true)
			})
			if i < n-1 {
				ui.Box(c).Height(1).Background(p.Separator).Margin(0, 0, 0, 44)
			}
		}
	})
}

func (a *app) callsWidget(c *ui.Context, p palette) {
	widget(c, p, icPhone, p.Green, "最近通话", func() {
		if len(a.calls) == 0 {
			ui.Text(c, "没有通话记录").TextColor(p.Secondary).FontSize(13)
			return
		}
		for _, r := range a.calls[:min(len(a.calls), 3)] {
			ui.Row(c.Key(r.ID)).Gap(8).AlignItems(ui.Center).Children(func() {
				ic := icOutgoing
				if r.Direction == "incoming" {
					ic = icIncoming
				}
				ui.Icon(c, ic).FontSize(14).TextColor(p.Secondary)
				name := ui.Text(c, displayNumber(r.Number)).FontSize(14).SingleLine().Grow(1)
				if r.Missed() {
					name.TextColor(p.Red)
				}
				ui.Text(c, relTime(r.StartedAt, a.now())).FontSize(12).TextColor(p.Secondary)
			})
		}
	})
}

func (a *app) tasksWidget(c *ui.Context, p palette) {
	var enabled []Task
	for _, t := range a.tasks {
		if t.Enabled {
			enabled = append(enabled, t)
		}
	}
	if len(enabled) == 0 {
		return
	}
	widget(c, p, icClock, p.Orange, "自动任务", func() {
		for _, t := range enabled[:min(len(enabled), 4)] {
			ui.Row(c.Key(t.ID)).Gap(8).Children(func() {
				ui.Text(c, t.Name).FontSize(14).SingleLine().Grow(1)
				ui.Text(c, untilTime(t.NextRunAt, a.now())).FontSize(12).TextColor(p.Secondary)
			})
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
