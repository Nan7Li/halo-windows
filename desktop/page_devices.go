package main

import (
	"context"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

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
	ui.Row(c).Grow(1).MinHeight(0).Gap(16).Padding(0, 16, 16, 16).AlignItems(ui.Stretch).Children(func() {
		ui.Scroll(c).Width(260).Shrink(0).Gap(8).Children(func() {
			for _, d := range a.devices {
				a.deviceCard(c, p, d)
			}
		})
		ui.Scroll(c).Grow(1).MinWidth(0).Gap(18).Children(func() {
			a.errorBanner(c, p, "devices")
			if d := a.deviceByID(a.deviceID); d != nil {
				a.deviceDetail(c, p, *d)
			}
		})
	})
}

func (a *app) deviceByID(id string) *Device {
	for i := range a.devices {
		if a.devices[i].ID == id {
			return &a.devices[i]
		}
	}
	return nil
}

func (a *app) deviceCard(c *ui.Context, p palette, d Device) {
	t := c.Theme()
	sel := d.ID == a.deviceID
	card := ui.Column(c.Key(d.ID)).Padding(12, 14).Gap(4).Radius(14).Background(p.Card).Cursor(ui.CursorPointer).Label(d.Title())
	if sel {
		card.Border(2, t.Accent)
	} else if card.Hovered() {
		card.Border(2, p.Separator)
	} else {
		card.Border(2, ui.Transparent)
	}
	card.Children(func() {
		status, tone := d.Status()
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, d.Title()).Bold().SingleLine().Grow(1)
			signalBars(c, p, d.Bars())
		})
		ui.Text(c, joinDot(d.Modem.Operator, d.Modem.NetworkMode, d.Modem.RadioBand)).FontSize(12).TextColor(p.Secondary).SingleLine()
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			dot(c, p.tone(tone))
			ui.Text(c, status).FontSize(12).TextColor(p.Secondary)
		})
	})
	if card.Clicked() {
		a.deviceID = d.ID
	}
}

func (a *app) deviceDetail(c *ui.Context, p palette, d Device) {
	status, tone := d.Status()
	ui.Row(c).Gap(14).AlignItems(ui.Center).Padding(4, 4, 0).Children(func() {
		iconTile(c, icDevices, p.Green)
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, d.Title()).FontSize(20).Bold().SingleLine()
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				dot(c, p.tone(tone))
				ui.Text(c, joinDot(status, d.Modem.Model)).FontSize(13).TextColor(p.Secondary)
			})
		})
	})
	busy := a.deviceBusy[d.ID]
	section(c, p, "控制", func() {
		row(c, p, false, func() {
			iconTile(c, icAirplane, p.Orange)
			ui.Text(c, "飞行模式").Grow(1)
			on := d.FlightMode || d.RadioMode == "airplane"
			if ui.Switch(c, &on).Label("飞行模式").Disabled(busy).Changed() {
				a.setDevice(d.ID, func(x *Device) { x.FlightMode = on }, map[bool]string{true: "飞行模式已开启", false: "飞行模式已关闭"}[on],
					func(ctx context.Context) error { return a.api.SetFlightMode(ctx, d.ID, on) })
			}
		})
		row(c, p, false, func() {
			iconTile(c, icWifi, p.Blue)
			ui.Column(c).Grow(1).Children(func() {
				ui.Text(c, "WiFi 通话")
				ui.Text(c, vowifiState(d)).FontSize(12).TextColor(p.Secondary)
			})
			on := d.VoWiFiEnabled
			if ui.Switch(c, &on).Label("WiFi 通话").Disabled(busy).Changed() {
				a.setDevice(d.ID, func(x *Device) { x.VoWiFiEnabled = on }, map[bool]string{true: "WiFi 通话已开启，等待 IMS 注册…", false: "WiFi 通话已关闭"}[on],
					func(ctx context.Context) error { return a.api.SetVoWiFi(ctx, d.ID, on) })
			}
		})
		if d.VoWiFiEnabled {
			row(c, p, false, func() {
				iconTile(c, icRefresh, p.Teal)
				ui.Text(c, "重新连接 WiFi 通话").Grow(1)
				if ui.Button(c, "重新连接").Disabled(busy).Clicked() {
					a.setDevice(d.ID, nil, "已触发重连，WiFi 通话正在重启", func(ctx context.Context) error { return a.api.ReconnectVoWiFi(ctx, d.ID) })
				}
			})
		}
		row(c, p, true, func() {
			iconTile(c, icPower, p.Red)
			ui.Text(c, "重启模组").Grow(1)
			if ui.Button(c, "重启…").Disabled(busy).Clicked() {
				a.rebootAsk = true
			}
		})
	})
	if ui.AlertDialog(c, &a.rebootAsk, "重启“"+d.Title()+"”？", "通话和数据会中断约一分钟。", "取消", "重启") == 1 {
		a.setDevice(d.ID, func(x *Device) { x.ModemPhase = "rebooting" }, "重启指令已送达", func(ctx context.Context) error { return a.api.Reboot(ctx, d.ID) })
	}
	kv := func(rows [][2]string) {
		var shown [][2]string
		for _, r := range rows {
			if r[1] != "" && r[1] != "—" {
				shown = append(shown, r)
			}
		}
		for i, r := range shown {
			row(c, p, i == len(shown)-1, func() {
				ui.Text(c, r[0]).Grow(1)
				ui.Text(c, r[1]).TextColor(p.Secondary).Selectable().SingleLine()
			})
		}
	}
	section(c, p, "蜂窝网络", func() {
		signal := ""
		if d.Modem.SignalDbm != 0 {
			signal = fmt.Sprintf("%.0f dBm", d.Modem.SignalDbm)
			if d.Modem.SignalSinr != 0 {
				signal += fmt.Sprintf(" · SINR %.0f dB", d.Modem.SignalSinr)
			}
		}
		kv([][2]string{
			{"运营商", d.Modem.Operator},
			{"网络", joinDot(d.Modem.NetworkMode, d.Modem.RadioBand)},
			{"信号", signal},
			{"注册", registrationName(d.Registration)},
			{"公网 IP", d.PublicIP},
			{"本机号码", d.VoWiFi.LocalPhone},
		})
	})
	section(c, p, "识别信息", func() {
		kv([][2]string{
			{"型号", d.Modem.Model},
			{"固件", d.Modem.Firmware},
			{"IMEI", d.Modem.IMEI},
			{"ICCID", d.Modem.ICCID},
			{"IMSI", d.Modem.IMSI},
		})
	})
	if d.VoWiFi.LastError != "" {
		section(c, p, "WiFi 通话最近的错误", func() {
			row(c, p, true, func() {
				ui.Text(c, d.VoWiFi.LastError).FontSize(12).TextColor(p.Secondary).Selectable().Grow(1)
			})
		})
	}
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
	case d.VoWiFi.IMSReady:
		return "IMS 已注册，可以通话和收发短信"
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
