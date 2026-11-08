package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// devicesPage is the list of devices beside the one chosen, as the web
// UI's devices page on the desktop.
func (a *app) devicesPage(c *ui.Context, p palette) {
	if !a.loadState(c, p, "devices") {
		return
	}
	if len(a.devices) == 0 {
		emptyState(c, p, icDevices, "还没有设备", "在 Halo 网页的“设备”里添加模组")
		return
	}
	if a.deviceID == "" || a.deviceByID(a.deviceID) == nil {
		a.deviceID = a.devices[0].ID
	}
	ui.Row(c).Grow(1).MinHeight(0).Gap(24).Padding(4, 28, 0, 28).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Width(320).Shrink(0).Gap(10).Children(func() {
			ui.SearchField(c, &a.deviceQuery).Placeholder("搜索设备 / ICCID / IMEI").Label("搜索设备").Radius(9).Background(p.Fill).Border(0, ui.Transparent)
			list := a.filteredDevices()
			ui.Column(c).Radius(12).Background(p.Card).Clip().Children(func() {
				for i, d := range list {
					a.deviceRow(c, p, d, i == len(list)-1)
				}
			})
			ui.Textf(c, "共 %d 台设备", len(a.devices)).FontSize(11).TextColor(p.Secondary).Padding(0, 4)
		})
		ui.Scroll(c).Grow(1).MinWidth(0).Padding(0, 0, 28).Gap(16).Children(func() {
			a.errorBanner(c, p, "devices")
			if d := a.deviceByID(a.deviceID); d != nil {
				a.deviceDetail(c, p, *d)
			}
		})
	})
}

func (a *app) filteredDevices() []Device {
	q := strings.ToLower(strings.TrimSpace(a.deviceQuery))
	if q == "" {
		return a.devices
	}
	var out []Device
	for _, d := range a.devices {
		if strings.Contains(strings.ToLower(d.Title()+" "+d.ID+" "+d.Modem.ICCID+" "+d.Modem.IMEI+" "+d.Modem.Operator), q) {
			out = append(out, d)
		}
	}
	return out
}

func (a *app) deviceByID(id string) *Device {
	for i := range a.devices {
		if a.devices[i].ID == id {
			return &a.devices[i]
		}
	}
	return nil
}

func (a *app) deviceRow(c *ui.Context, p palette, d Device, last bool) {
	t := c.Theme()
	sel := d.ID == a.deviceID
	r := ui.Row(c.Key(d.ID)).MinHeight(52).Padding(8, 12).Gap(10).AlignItems(ui.Center).Cursor(ui.CursorPointer).Role(ui.RoleTab).AccessSelected(sel).Label(d.Title())
	fg, sub := p.Label, p.Secondary
	if sel {
		r.Background(t.Accent)
		fg, sub = ui.RGB(255, 255, 255), ui.RGBA(255, 255, 255, 0.8)
	} else if r.Hovered() {
		r.Background(p.Fill.Alpha(0.06))
	}
	r.Children(func() {
		ui.Box(c).Size(30, 30).Radius(8).Center().Shrink(0).Background(p.Fill).Children(func() {
			ui.Icon(c, icDevices).FontSize(18).TextColor(fg)
		})
		status, tone := d.Status()
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, d.Title()).Bold().SingleLine().TextColor(fg)
			ui.Text(c, joinDot(d.Modem.Operator, d.Modem.NetworkMode)).FontSize(11).TextColor(sub).SingleLine()
		})
		ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
			col := p.tone(tone)
			if sel {
				col = ui.RGBA(255, 255, 255, 0.85)
			}
			dot(c, col)
			ui.Text(c, shortStatus(status)).FontSize(11).TextColor(map[bool]color{true: ui.RGB(255, 255, 255), false: col}[sel])
		})
		if !last && !sel {
			ui.Box(c).Absolute().Left(52).Right(0).Bottom(0).Height(1).Background(p.Separator)
		}
	})
	if r.Clicked() {
		a.deviceID = d.ID
	}
}

func shortStatus(s string) string {
	switch s {
	case "WiFi 通话就绪":
		return "在线"
	}
	return s
}

// actionTile is one of the device's controls: an icon over a label, filled
// in its color while what it controls is on.
func actionTile(c *ui.Context, p palette, ic *ui.SVG, label string, col color, on, busy bool) ui.Element {
	b := ui.Column(c.Key(label)).Grow(1).Basis(0).MinWidth(0).Height(54).Radius(12).Gap(3).Center().Cursor(ui.CursorPointer).
		Role(ui.RoleButton).Label(label).Disabled(busy)
	fg := col
	if on {
		b.Background(col)
		fg = ui.RGB(255, 255, 255)
	} else {
		b.Background(p.Card)
		if b.Pressed() {
			b.Background(p.FillPressed)
		}
	}
	return b.Children(func() {
		ui.Icon(c, ic).FontSize(18).TextColor(fg)
		ui.Text(c, label).FontSize(11).TextColor(fg).SingleLine()
	})
}

func (a *app) deviceDetail(c *ui.Context, p palette, d Device) {
	busy := a.deviceBusy[d.ID]
	ui.Row(c).Gap(12).AlignItems(ui.Center).Padding(0, 4).Children(func() {
		ui.Box(c).Size(52, 52).Radius(13).Center().Background(p.Card).Shadow(0, 1, 3, 0, ui.RGBA(0, 0, 0, 0.08)).Children(func() {
			ui.Icon(c, icDevices).FontSize(28).TextColor(p.Secondary)
		})
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, d.Title()).FontSize(17).Bold().SingleLine()
			kind := "蜂窝"
			if d.VoWiFiEnabled {
				kind = "VoWiFi"
			}
			ui.Text(c, joinDot(d.Modem.Operator, kind, d.ID)).FontSize(12).TextColor(p.Secondary).SingleLine()
		})
	})
	flight := d.FlightMode || d.RadioMode == "airplane"
	ui.Row(c).Gap(10).Children(func() {
		if actionTile(c, p, icWifi, "VoWiFi", p.Blue, d.VoWiFiEnabled, busy).Clicked() {
			on := !d.VoWiFiEnabled
			a.setDevice(d.ID, func(x *Device) { x.VoWiFiEnabled = on }, map[bool]string{true: "WiFi 通话已开启，等待 IMS 注册…", false: "WiFi 通话已关闭"}[on],
				func(ctx context.Context) error { return a.api.SetVoWiFi(ctx, d.ID, on) })
		}
		if actionTile(c, p, icAirplane, "飞行模式", p.Orange, flight, busy).Clicked() {
			on := !flight
			a.setDevice(d.ID, func(x *Device) {
				x.FlightMode = on
				if !on && x.RadioMode == "airplane" {
					x.RadioMode = "cellular"
				}
			}, map[bool]string{true: "飞行模式已开启", false: "飞行模式已关闭"}[on],
				func(ctx context.Context) error { return a.api.SetFlightMode(ctx, d.ID, on) })
		}
		if actionTile(c, p, icRefresh, "重连 VoWiFi", p.Blue, false, busy || !d.VoWiFiEnabled).Clicked() {
			a.setDevice(d.ID, nil, "已触发重连，WiFi 通话正在重启", func(ctx context.Context) error { return a.api.ReconnectVoWiFi(ctx, d.ID) })
		}
		if actionTile(c, p, icMessages, "短信", p.Blue, false, false).Clicked() {
			a.search = d.Title()
			a.goTo(pageMessages)
		}
		if actionTile(c, p, icPower, "重启模组", p.Red, false, busy).Clicked() {
			a.rebootAsk = true
		}
	})
	if ui.AlertDialog(c, &a.rebootAsk, "重启“"+d.Title()+"”？", "通话和数据会中断约一分钟。", "取消", "重启") == 1 {
		a.setDevice(d.ID, func(x *Device) { x.ModemPhase = "rebooting" }, "重启指令已送达", func(ctx context.Context) error { return a.api.Reboot(ctx, d.ID) })
	}
	card := func(title string, fn func()) ui.Element {
		return ui.Column(c).Grow(1).Basis(0).MinWidth(0).Padding(14, 16).Gap(10).Radius(14).Background(p.Card).Children(func() {
			ui.Text(c, title).FontSize(11).Bold().TextColor(p.Secondary)
			fn()
		})
	}
	kv := func(rows [][2]string) {
		ui.Column(c).Gap(7).Children(func() {
			for _, r := range rows {
				if r[1] == "" || r[1] == "—" {
					continue
				}
				ui.Row(c).Gap(12).Children(func() {
					ui.Text(c, r[0]).TextColor(p.Secondary).Shrink(0)
					ui.Text(c, r[1]).Selectable().SingleLine().TextAlign(ui.End).Grow(1).MinWidth(0)
				})
			}
		})
	}
	ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
		card("运行状态", func() {
			status, tone := d.Status()
			col := p.tone(tone)
			text := status
			if d.VoWiFiEnabled {
				text = "WiFi-Calling · " + vowifiState(d)
			}
			ui.Row(c).Padding(10, 14).Gap(8).Radius(10).AlignItems(ui.Center).Background(col.Alpha(0.12)).Children(func() {
				dot(c, col)
				ui.Text(c, text).Bold().TextColor(col).Grow(1).MaxLines(2)
			})
			if d.VoWiFiEnabled {
				steps := []struct {
					label string
					ok    bool
				}{{"SIM", d.VoWiFi.SIMReady}, {"Access", d.VoWiFi.AccessReady}, {"Tunnel", d.VoWiFi.TunnelReady}, {"IMS", d.VoWiFi.IMSReady}, {"SMS", d.VoWiFi.SMSReady}}
				ui.Row(c).Gap(4).Children(func() {
					for _, s := range steps {
						ui.Column(c).Grow(1).Basis(0).Gap(5).AlignItems(ui.Center).Children(func() {
							fill := p.Green
							if !s.ok {
								fill = p.Fill
							}
							ui.Box(c).Height(5).AlignSelf(ui.Stretch).Radius(3).Background(fill)
							ui.Text(c, s.label).FontSize(10).TextColor(p.Secondary)
						})
					}
				})
			}
			if d.VoWiFi.LastError != "" {
				ui.Text(c, d.VoWiFi.LastError).FontSize(11).TextColor(p.Secondary).Selectable().MaxLines(3)
			}
			signal := ""
			if d.Modem.SignalDbm != 0 {
				signal = fmt.Sprintf("%.0f dBm", d.Modem.SignalDbm)
				if d.Modem.SignalSinr != 0 {
					signal += fmt.Sprintf(" · SINR %.0f dB", d.Modem.SignalSinr)
				}
			}
			kv([][2]string{
				{"网络", joinDot(d.Modem.NetworkMode, d.Modem.RadioBand)},
				{"信号", signal},
				{"注册", registrationName(d.Registration)},
				{"公网 IP", d.PublicIP},
			})
		})
		card("SIM / 设备", func() {
			kv([][2]string{
				{"IMEI", d.Modem.IMEI},
				{"ICCID", d.Modem.ICCID},
				{"IMSI", d.Modem.IMSI},
				{"本机号码", d.VoWiFi.LocalPhone},
				{"运营商", d.Modem.Operator},
				{"型号", d.Modem.Model},
				{"固件版本", d.Modem.Firmware},
				{"飞行模式", map[bool]string{true: "是", false: "否"}[flight]},
			})
		})
	})
}

// setDevice changes a device at once, sends the change, then reloads the
// devices, which puts back the old state if the change failed.
func (a *app) setDevice(id string, change func(*Device), done string, work func(ctx context.Context) error) {
	if change != nil {
		if d := a.deviceByID(id); d != nil {
			change(d)
		}
	}
	if a.api.Preview() {
		a.say(done + "（预览，未发送到服务器）")
		return
	}
	a.deviceBusy[id] = true
	a.spawn(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := work(ctx)
		a.update(func() {
			a.deviceBusy[id] = false
			if err != nil {
				a.say(errText(err))
			} else if done != "" {
				a.say(done)
			}
			a.loadDevices()
		})
	})
}

func vowifiState(d Device) string {
	switch {
	case !d.VoWiFiEnabled:
		return "已关闭"
	case d.VoWiFi.IMSReady && d.VoWiFi.SMSReady:
		return "全部就绪"
	case d.VoWiFi.IMSReady:
		return "IMS 已注册"
	case d.VoWiFi.Phase != "":
		return "正在连接（" + d.VoWiFi.Phase + "）"
	}
	return "正在连接"
}

func registrationName(s string) string {
	switch s {
	case "registered":
		return "已注册"
	case "searching":
		return "正在搜索"
	case "denied":
		return "被拒绝"
	}
	return ""
}
