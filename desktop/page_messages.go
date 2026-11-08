package main

import (
	"strings"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// messagesPage is Messages for the Mac: conversations beside the one open.
func (a *app) messagesPage(c *ui.Context, p palette) {
	ui.Row(c).Grow(1).MinHeight(0).Margin(sidebarInset, sidebarInset, sidebarInset, sidebarInset).Radius(14).Background(p.Card).Clip().AlignItems(ui.Stretch).Children(func() {
		a.conversationList(c, p)
		ui.Box(c).Width(1).Background(p.Separator)
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			switch {
			case a.composeOpen:
				a.composeView(c, p)
			case a.selected != "":
				a.conversation(c, p)
			default:
				emptyState(c, p, icMessages, "选择一个对话", "短信由 Halo 上的模组收发")
			}
		})
	})
}

func (a *app) filteredContacts() []Contact {
	q := strings.ToLower(strings.TrimSpace(a.search))
	if q == "" {
		return a.contacts
	}
	var out []Contact
	for _, k := range a.contacts {
		if strings.Contains(strings.ToLower(k.Title()+" "+k.Peer+" "+k.LastMessage), q) {
			out = append(out, k)
		}
	}
	return out
}

func (a *app) conversationList(c *ui.Context, p palette) {
	t := c.Theme()
	list := a.filteredContacts()
	ui.Column(c).Width(320).Shrink(0).Children(func() {
		bar := c.TitleBar()
		ui.Row(c).Height(titleBarHeight-sidebarInset).Shrink(0).Padding(0, 10, 0, 16).Gap(8).AlignItems(ui.Center).DragWindow().Children(func() {
			ui.Text(c, "信息").FontSize(15).Bold().Grow(1)
			if a.refreshing["contacts"] && a.loaded["contacts"] {
				ui.Spinner(c).Size(14, 14)
			}
			if glassIconButton(c, icRefresh, "刷新").Clicked() {
				a.refreshPage()
			}
			if glassIconButton(c, icCompose, "新信息").Clicked() {
				a.composeOpen, a.selected, a.draft = true, "", ""
				if lines := a.smsLines(); len(lines) > 0 && a.composeLine == "" {
					a.composeLine = lines[0].ID
				}
			}
		})
		_ = bar
		ui.Row(c).Padding(2, 12, 8).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.SearchField(c, &a.search).Placeholder("搜索").Grow(1).Label("搜索短信").Radius(9).Background(p.Fill).Border(0, ui.Transparent)
		})
		if !a.loadState(c, p, "contacts") {
			return
		}
		ui.List(c, nil, len(list), func(i int) {
			k := list[i]
			key := k.Key()
			sel := key == a.selected && !a.composeOpen
			r := ui.Row(c.Key(key)).Padding(9, 10, 9, 6).Margin(0, 8).Gap(8).Radius(10).AlignItems(ui.Center).Cursor(ui.CursorPointer)
			if sel {
				r.Background(t.Accent)
			} else if r.Hovered() {
				r.Background(p.Fill.Alpha(0.06))
			}
			fg, sub := p.Label, p.Secondary
			if sel {
				fg, sub = ui.RGB(255, 255, 255), ui.RGBA(255, 255, 255, 0.82)
			}
			r.Children(func() {
				ui.Box(c).Width(8).Shrink(0).Center().Children(func() {
					if k.UnreadCount > 0 && !sel {
						ui.Box(c).Size(8, 8).Radius(4).Background(t.Accent)
					}
				})
				avatar(c, p, k.Title(), 40)
				ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Text(c, k.Title()).Bold().SingleLine().Grow(1).TextColor(fg)
						ui.Text(c, relTime(k.LastTime, a.now())).FontSize(12).TextColor(sub)
					})
					ui.Text(c, k.LastMessage).FontSize(12).TextColor(sub).MaxLines(2)
					if k.DeviceName != "" && len(a.devices) > 1 {
						ui.Text(c, joinDot(k.DeviceName, k.LocalPhone)).FontSize(11).TextColor(sub).SingleLine()
					}
				})
			})
			if r.Clicked() {
				a.composeOpen = false
				a.openThread(key, true)
			}
		}).Grow(1).Padding(0, 0, 8).Children(func() {
			if len(list) == 0 {
				msg := "没有短信"
				if a.search != "" {
					msg = "没有找到“" + a.search + "”"
				}
				ui.Text(c, msg).TextColor(p.Secondary).FontSize(13).Padding(24).TextAlign(ui.Center)
			}
		})
	})
}

func (a *app) conversation(c *ui.Context, p palette) {
	k, ok := a.contactByKey(a.selected)
	if !ok {
		emptyState(c, p, icMessages, "对话不存在", "")
		return
	}
	t := c.Theme()
	ui.Column(c).Grow(1).MinHeight(0).Children(func() {
		ui.Box(c).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(p.Separator).DragWindow().Children(func() {
			ui.Column(c).Padding(10, 0, 10).Gap(4).AlignItems(ui.Center).Children(func() {
				avatar(c, p, k.Title(), 40)
				ui.Text(c, k.Title()).FontSize(11).Padding(1, 8).Radius(9).Background(p.Fill).SingleLine()
				sub := k.Peer
				if k.Title() == k.Peer {
					sub = ""
				}
				if s := joinDot(sub, k.DeviceName); s != "" {
					ui.Text(c, s).FontSize(11).TextColor(p.Secondary).SingleLine()
				}
			})
			if isDialable(k.Peer) {
				ui.Box(c).Absolute().Top(0).Bottom(0).Right(c.TitleBar().Right + 14).Center().Children(func() {
					if glassIconButton(c, icPhone, "呼叫 "+k.Peer).Clicked() {
						a.dialLine = k.DeviceID
						a.dial(k.Peer)
					}
				})
			}
		})
		ui.Box(c).Grow(1).MinHeight(0).Children(func() {
			msgs := a.thread
			if a.threadKey != a.selected {
				msgs = nil
			}
			list := ui.List(c.Key("thread:"+a.selected), &a.threadList, len(msgs), func(i int) {
				m := msgs[i]
				showTime := i == 0 || m.Timestamp.Sub(msgs[i-1].Timestamp) > time.Hour
				ui.Column(c.Key(m.ID)).Gap(4).Padding(1, 16).Children(func() {
					if showTime {
						ui.Text(c, bubbleTime(m.Timestamp, a.now())).FontSize(11).TextColor(p.Secondary).AlignSelf(ui.Center).Padding(10, 0, 4)
					}
					mine := m.Mine()
					// The last bubble of a run from one side has a tail.
					tail := i == len(msgs)-1 || msgs[i+1].Mine() != mine || msgs[i+1].Timestamp.Sub(m.Timestamp) > time.Hour
					fill, fg := p.BubbleOther, p.Label
					if mine {
						fill, fg = t.Accent, ui.RGB(255, 255, 255)
					}
					ui.Row(c).Justify(map[bool]ui.Align{true: ui.End, false: ui.Start}[mine]).Padding(0, 6).Children(func() {
						bub := ui.Text(c, m.Text()).Selectable().Padding(7, 12).Radius(17).MaxWidthPercent(70).Background(fill).TextColor(fg)
						if tail {
							bub.Draw(func(pt *ui.Painter, r ui.Rect) { bubbleTail(pt, r, mine, fill) })
						}
						bub.Tooltip(m.Timestamp.Local().Format("2006-01-02 15:04:05"))
					})
					if m.Mine() && i == len(msgs)-1 {
						ui.Text(c, deliveryText(m)).FontSize(11).TextColor(p.Secondary).AlignSelf(ui.End)
					}
				})
			}).Fill().Padding(8, 0, 76)
			list.Children(func() {
				if len(msgs) == 0 && a.refreshing["thread"] {
					ui.Column(c).Grow(1).Center().Children(func() { ui.Spinner(c) })
				}
			})
			a.composer(c, p, func() { a.sendMessage() }, "短信")
		})
	})
}

// composer is the message field, a capsule of glass floating over the
// bottom of the conversation.
func (a *app) composer(c *ui.Context, p palette, send func(), placeholder string) {
	t := c.Theme()
	ui.Row(c).Absolute().Left(14).Right(14).Bottom(14).Padding(6, 6, 6, 16).Gap(8).AlignItems(ui.End).Radius(22).
		Material(glass.Glass{}).Children(func() {
		ui.TextAreaBase(c.Key("draft"), &a.draft).Lines(1, 6).Placeholder(placeholder).PlaceholderColor(p.Tertiary).
			Label("短信内容").FontSize(14).Grow(1).MinWidth(0).Padding(5, 0).
			OnTextKey(func(k ui.TextKey) bool {
				if k.Key == ui.KeyEnter && k.Mods == 0 {
					a.draft = k.Text()
					send()
					k.SetText(a.draft)
					return true
				}
				return false
			})
		ok := strings.TrimSpace(a.draft) != "" && !a.sending
		fill := t.Accent
		if !ok {
			fill = p.Tertiary
		}
		b := ui.Box(c).Size(30, 30).Radius(15).Center().Background(fill).Cursor(ui.CursorPointer).Role(ui.RoleButton).Label("发送").Disabled(!ok)
		b.Children(func() {
			ui.Icon(c, icUp).FontSize(17).TextColor(ui.RGB(255, 255, 255))
		})
		if b.Clicked() && ok {
			send()
		}
	})
}

func (a *app) smsLines() []Device {
	var out []Device
	for _, d := range a.devices {
		if d.SMSEnabled || d.VoWiFi.SMSReady {
			out = append(out, d)
		}
	}
	if len(out) == 0 {
		return a.devices
	}
	return out
}

func (a *app) composeView(c *ui.Context, p palette) {
	lines := a.smsLines()
	ui.Column(c).Grow(1).MinHeight(0).Children(func() {
		ui.Row(c).Padding(12, 18).Gap(10).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(p.Separator).Children(func() {
			ui.Text(c, "收件人：").TextColor(p.Secondary)
			ui.TextInput(c.Key("to"), &a.composeTo).Placeholder("电话号码").Label("收件人").AutoFocus().Grow(1).
				Background(ui.Transparent).Border(0, ui.Transparent)
			if len(lines) > 1 {
				names := make([]string, len(lines))
				chosen := ""
				for i, d := range lines {
					names[i] = d.Title()
					if d.ID == a.composeLine {
						chosen = d.Title()
					}
				}
				if ui.Select(c, &chosen, names).Label("发送线路").Width(170).Changed() {
					for _, d := range lines {
						if d.Title() == chosen {
							a.composeLine = d.ID
						}
					}
				}
			}
			if ui.Button(c, "取消").Clicked() {
				a.composeOpen = false
			}
		})
		ui.Box(c).Grow(1).Children(func() {
			emptyState(c, p, icCompose, "新信息", "输入号码和内容，按回车发送")
			a.composer(c, p, func() { a.sendNew() }, "短信")
		})
	})
}

// bubbleTail draws the tail at a bubble's bottom corner, on the side of
// who sent it, as Messages does.
func bubbleTail(pt *ui.Painter, r ui.Rect, right bool, fill color) {
	b := r.Y + r.H
	path := &ui.Path{}
	if right {
		e := r.X + r.W
		path.MoveTo(e-16, b).LineTo(e+5, b).QuadTo(e-1, b-3, e-1, b-15).LineTo(e-16, b-15).Close()
	} else {
		e := r.X
		path.MoveTo(e+16, b).LineTo(e-5, b).QuadTo(e+1, b-3, e+1, b-15).LineTo(e+16, b-15).Close()
	}
	pt.FillPath(path, fill)
}

func deliveryText(m Message) string {
	switch strings.ToLower(m.DeliveryState + m.Status) {
	case "sending":
		return "正在发送…"
	case "delivered":
		return "已送达"
	case "failed":
		return "未能发送"
	}
	if strings.Contains(strings.ToLower(m.DeliveryState), "fail") || strings.Contains(strings.ToLower(m.Status), "fail") {
		return "未能发送"
	}
	return "已发送"
}

func bubbleTime(t, now time.Time) string {
	rel := relTime(t, now)
	clock := t.Local().Format("15:04")
	if rel == clock {
		return "今天 " + clock
	}
	return rel + " " + clock
}

func isDialable(s string) bool {
	digits := 0
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+' || r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			return false
		}
	}
	return digits >= 3
}
