package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

// ---- Network: WireGuard tunnels, as iOS's VPN settings ----

func (a *app) networkPage(c *ui.Context, p palette) {
	if !a.loadState(c, p, "tunnels") {
		return
	}
	pageScroll(c, func() {
		a.errorBanner(c, p, "tunnels")
		snap := a.tunnels
		if !snap.Available && len(snap.Tunnels) == 0 {
			emptyState(c, p, icTunnel, "WireGuard 不可用", nonEmpty(snap.Hint, "Halo 主机上没有安装 WireGuard"))
			return
		}
		if len(snap.Tunnels) == 0 {
			emptyState(c, p, icTunnel, "没有隧道", "在 Halo 网页的“网络 › WireGuard”里添加")
			return
		}
		section(c, p, "WireGuard 隧道", func() {
			for i, tn := range snap.Tunnels {
				row(c.Key(tn.ID), p, i == len(snap.Tunnels)-1, func() {
					fill := p.Gray
					if tn.Running {
						fill = p.Green
					}
					iconTile(c, icTunnel, fill)
					ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
						ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
							ui.Text(c, nonEmpty(tn.Name, tn.ID)).Bold().SingleLine()
							if tn.External {
								ui.Badge(c, "外部")
							}
						})
						ui.Text(c, tunnelDetail(tn)).FontSize(12).TextColor(p.Secondary).SingleLine()
						if tn.Error != "" {
							ui.Text(c, tn.Error).FontSize(12).TextColor(p.Red).MaxLines(2)
						}
					})
					on := tn.Running
					sw := ui.Switch(c, &on).Label("连接 " + tn.Name).Disabled(a.deviceBusy["wg:"+tn.ID] || tn.External || !tn.Available)
					if sw.Changed() {
						a.setTunnel(tn, on)
					}
				})
			}
		})
	})
}

func tunnelDetail(t Tunnel) string {
	state := "未连接"
	if t.Running {
		state = "已连接"
		if t.TransferRx+t.TransferTx > 0 {
			state += fmt.Sprintf(" · ↓%s ↑%s", formatBytes(float64(t.TransferRx)), formatBytes(float64(t.TransferTx)))
		}
	}
	return joinDot(state, t.Interface, t.Endpoint)
}

func (a *app) setTunnel(t Tunnel, up bool) {
	for i := range a.tunnels.Tunnels {
		if a.tunnels.Tunnels[i].ID == t.ID {
			a.tunnels.Tunnels[i].Running = up
		}
	}
	done := "隧道已断开"
	if up {
		done = "隧道已连接"
	}
	if a.api.Preview() {
		a.say(done + "（预览，未发送到服务器）")
		return
	}
	key := "wg:" + t.ID
	a.deviceBusy[key] = true
	a.spawn(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30e9)
		defer cancel()
		err := a.api.SetTunnel(ctx, t.ID, up)
		a.update(func() {
			a.deviceBusy[key] = false
			if err != nil {
				a.say(errText(err))
			} else {
				a.say(done)
			}
			a.loadTunnels()
		})
	})
}

// ---- Logs ----

func (a *app) logsPage(c *ui.Context, p palette) {
	ui.Row(c).Padding(0, 22, 10).Gap(10).AlignItems(ui.Center).Children(func() {
		if ui.Segmented(c, &a.logLevel, "全部", "信息", "警告", "错误").Label("级别").Changed() {
			a.loadLogs()
		}
		ui.Box(c).Grow(1)
		ui.SearchField(c, &a.logQuery).Placeholder("筛选").Width(220).Label("筛选日志")
	})
	if !a.loadState(c, p, "logs") {
		return
	}
	q := strings.ToLower(strings.TrimSpace(a.logQuery))
	var list []LogEntry
	for _, e := range a.logs {
		if q == "" || strings.Contains(strings.ToLower(e.Message), q) {
			list = append(list, e)
		}
	}
	ui.Column(c).Grow(1).MinHeight(0).Margin(0, 16, 16, 16).Radius(14).Background(p.Card).Clip().Children(func() {
		ui.List(c, &a.logList, len(list), func(i int) {
			e := list[i]
			ui.Row(c).Padding(5, 14).Gap(10).AlignItems(ui.Start).Children(func() {
				ui.Text(c, e.Time.Local().Format("01-02 15:04:05")).FontSize(12).TextColor(p.Secondary).FontFeatures("tnum").Shrink(0)
				lv := strings.ToUpper(e.Level)
				if len(lv) > 5 {
					lv = lv[:5]
				}
				ui.Text(c, lv).FontSize(11).Bold().TextColor(levelTone(p, e.Level)).Width(44).Shrink(0)
				msg := e.Message
				if len(e.Fields) > 0 {
					var b strings.Builder
					for k, v := range e.Fields {
						fmt.Fprintf(&b, " %s=%v", k, v)
					}
					msg += b.String()
				}
				ui.Text(c, msg).FontSize(12).Selectable().Grow(1).MinWidth(0)
			})
		}).Grow(1).Padding(6, 0).Children(func() {
			if len(list) == 0 {
				emptyState(c, p, icLogs, "没有日志", "")
			}
		})
	})
}

func levelTone(p palette, level string) color {
	switch strings.ToLower(level) {
	case "error":
		return p.Red
	case "warn", "warning":
		return p.Orange
	case "debug":
		return p.Gray
	}
	return p.Blue
}

// ---- Settings ----

func (a *app) settingsPage(c *ui.Context, p palette) {
	t := c.Theme()
	pageScroll(c, func() {
		ui.Column(c).Gap(18).Children(func() {
			section(c, p, "账户", func() {
				row(c, p, false, func() {
					ui.Avatar(c, a.user, nil).Size(40, 40)
					ui.Column(c).Grow(1).Children(func() {
						ui.Text(c, a.user).Bold()
						server := a.prefs.Server
						if a.api.Preview() {
							server = "预览模式，没有连接服务器"
						}
						ui.Text(c, server).FontSize(12).TextColor(p.Secondary).Selectable()
					})
				})
				if !a.api.Preview() {
					row(c, p, false, func() {
						iconTile(c, icExternal, p.Blue)
						ui.Text(c, "在浏览器中打开 Halo").Grow(1)
						ui.Link(c, "打开", a.prefs.Server)
					})
				}
				row(c, p, true, func() {
					iconTile(c, icSignOut, p.Red)
					label := "退出登录"
					if a.api.Preview() {
						label = "退出预览"
					}
					b := ui.Text(c, label).TextColor(p.Red).Grow(1).Cursor(ui.CursorPointer)
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

			section(c, p, "外观", func() {
				row(c, p, false, func() {
					iconTile(c, icPaint, p.Indigo)
					ui.Text(c, "外观").Grow(1)
					modes := []string{"", "light", "dark"}
					sel := 0
					for i, m := range modes {
						if m == a.prefs.Appearance {
							sel = i
						}
					}
					if ui.Segmented(c, &sel, "自动", "浅色", "深色").Label("外观").Changed() {
						a.prefs.Appearance = modes[sel]
						a.store.Save(a.prefs)
					}
				})
				row(c, p, false, func() {
					iconTile(c, icPaint, t.Accent)
					ui.Text(c, "强调色").Grow(1)
					ui.Row(c).Gap(8).Children(func() {
						for _, ac := range accents {
							fill := ac.light
							if p.dark {
								fill = ac.dark
							}
							sw := ui.Box(c.Key(ac.id)).Size(22, 22).Radius(11).Cursor(ui.CursorPointer).Role(ui.RoleRadio).AccessSelected(ac.id == a.prefs.Accent).Label(ac.name).Tooltip(ac.name)
							if ac.id == "" {
								sw.ConicGradient(ui.ConicGradient{Stops: []ui.GradientStop{
									{Color: p.Red, At: 0}, {Color: p.Orange, At: 0.2}, {Color: p.Green, At: 0.45}, {Color: p.Blue, At: 0.7}, {Color: p.Indigo, At: 0.85}, {Color: p.Red, At: 1},
								}})
							} else {
								sw.Background(fill)
							}
							if ac.id == a.prefs.Accent {
								sw.Border(2, p.Card).Shadow(0, 0, 0, 2, p.Label.Alpha(0.5))
							}
							if sw.Clicked() {
								a.prefs.Accent = ac.id
								a.store.Save(a.prefs)
							}
						}
					})
				})
				row(c, p, true, func() {
					iconTile(c, icSettings, p.Teal)
					ui.Column(c).Grow(1).Children(func() {
						ui.Text(c, "侧栏玻璃")
						ui.Text(c, "从磨砂到通透").FontSize(12).TextColor(p.Secondary)
					})
					if ui.Slider(c, &a.prefs.Clarity, 0, 1).Width(200).Label("侧栏玻璃通透度").Changed() {
						a.store.Save(a.prefs)
					}
				})
			})

			section(c, p, "通知", func() {
				row(c, p, true, func() {
					iconTile(c, icBell, p.Red)
					ui.Column(c).Grow(1).Children(func() {
						ui.Text(c, "新短信和来电")
						ui.Text(c, "客户端打开时用系统通知提醒").FontSize(12).TextColor(p.Secondary)
					})
					if ui.Switch(c, &a.prefs.Notify).Label("新短信和来电通知").Changed() {
						a.store.Save(a.prefs)
					}
				})
			})

			section(c, p, "关于", func() {
				row(c, p, false, func() {
					iconTile(c, icInfo, p.Gray)
					ui.Text(c, "Halo Desktop").Grow(1)
					ui.Text(c, appVersion).TextColor(p.Secondary)
				})
				row(c, p, false, func() {
					iconTile(c, icServer, p.Gray)
					ui.Text(c, "服务器版本").Grow(1)
					ui.Text(c, nonEmpty(a.version, "—")).TextColor(p.Secondary).Selectable()
				})
				row(c, p, true, func() {
					iconTile(c, icHalo, p.Gray)
					ui.Text(c, "源代码").Grow(1)
					ui.Link(c, "GitHub", "https://github.com/Nan7Li/halo-windows")
				})
			})
			ui.Text(c, "用 MyGo 原生绘制，同一套 Go 代码构建 Windows、macOS 和 Linux 版本。").FontSize(12).TextColor(p.Secondary).Padding(0, 12)
		})
	})
}
