package main

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

func (a *app) loginView(c *ui.Context, p palette) {
	t := c.Theme()
	wallpaper(c, p, ui.Box(c).Fill()).Children(func() {
		ui.Box(c).Absolute().Top(0).Left(0).Right(0).Height(max(c.TitleBar().Height, 36)).DragWindow()
		ui.Column(c).Fill().Center().Children(func() {
			card := ui.Column(c).Width(380).Padding(28, 28, 24).Gap(14).Radius(radiusWindow + 4).Material(glass.Glass{Clarity: 0.2, Tint: chrome.Tint})
			card.Children(func() {
				ui.Column(c).Gap(6).AlignItems(ui.Center).Padding(0, 0, 6).Children(func() {
					ui.Box(c).Size(56, 56).Radius(15).Center().Gradient(t.Accent, p.Indigo, 135).Children(func() {
						ui.Icon(c, icHalo).FontSize(32).TextColor(ui.RGB(255, 255, 255))
					})
					ui.Text(c, "登录 Halo").FontSize(22).Bold()
					ui.Text(c, "连接到你的 Halo 服务器").FontSize(13).TextColor(p.Secondary)
				})
				field := func(label string, fn func()) {
					ui.Column(c).Gap(4).Children(func() {
						ui.Text(c, label).FontSize(12).TextColor(p.Secondary)
						fn()
					})
				}
				submit := false
				field("服务器地址", func() {
					in := field27(ui.TextInput(c.Key("server"), &a.prefs.Server), p).Placeholder("例如 192.168.1.10:7575").Label("服务器地址")
					if a.prefs.Server == "" {
						in.AutoFocus()
					}
					submit = in.Submitted() || submit
				})
				field("用户名", func() {
					submit = field27(ui.TextInput(c.Key("user"), &a.prefs.Username), p).Placeholder("admin").Label("用户名").Submitted() || submit
				})
				field("密码", func() {
					in := field27(ui.TextInput(c.Key("password"), &a.password), p).Password().Label("密码")
					if a.prefs.Server != "" && a.prefs.Username != "" {
						in.AutoFocus()
					}
					submit = in.Submitted() || submit
				})
				ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "记住密码（保存在系统钥匙串）").FontSize(13).Grow(1)
					glassSwitch(c, p, &a.prefs.Remember).Label("记住密码")
				})
				if a.loginErr != "" {
					ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
						ui.Icon(c, icWarn).FontSize(14).TextColor(p.Red)
						ui.Text(c, a.loginErr).FontSize(12).TextColor(p.Red).Grow(1)
					})
				}
				ready := a.prefs.Server != "" && a.prefs.Username != "" && a.password != ""
				login := prominentButton(c, p, "登录", t.Accent, a.busy || !ready).AlignSelf(ui.Stretch)
				login.Children(func() {
					if a.busy {
						ui.Spinner(c).Size(14, 14).TextColor(ui.RGB(255, 255, 255))
						ui.Text(c, "正在登录…")
					} else {
						ui.Text(c, "登录")
					}
				})
				if (login.Clicked() || submit) && ready && !a.busy {
					server, user, pw := a.prefs.Server, a.prefs.Username, a.password
					a.busy = true
					a.spawn(func() { a.signIn(server, user, pw) })
				}
				if ui.Button(c, "不登录，先看看界面").Height(34).Radius(17).Background(ui.Transparent).Border(0, ui.Transparent).TextColor(t.Accent).Clicked() {
					a.enterPreview()
				}
			})
		})
	})
	if a.untrusted != nil {
		msg := "“" + a.untrusted.Subject + "”的证书不是受信任的机构签发的，可能是 Halo 自签名证书。\n\nSHA-256 指纹：\n" + groupHex(a.untrusted.Fingerprint) +
			"\n\n请在服务器的“设置 › HTTPS”里核对指纹一致后再信任。"
		switch ui.AlertDialog(c, &a.trustOpen, "要信任这台服务器吗？", msg, "取消", "信任并登录") {
		case 1:
			a.trust(a.prefs.Server, a.prefs.Username, a.password)
		case 0, ui.AlertDismissed:
			a.untrusted = nil
		}
	}
}

// field27 is a text field as iOS 27's forms draw one: a filled rounded
// rectangle, without a border.
func field27(e ui.Element, p palette) ui.Element {
	return e.Height(38).Radius(12).Padding(0, 12).Background(p.Fill).Border(0, ui.Transparent)
}

// groupHex writes a fingerprint in pairs, eight to a line.
func groupHex(h string) string {
	out := make([]byte, 0, len(h)*2)
	for i := 0; i+1 < len(h); i += 2 {
		if i > 0 {
			if i%16 == 0 {
				out = append(out, '\n')
			} else {
				out = append(out, ':')
			}
		}
		out = append(out, h[i]-32*b2i(h[i] >= 'a'), h[i+1]-32*b2i(h[i+1] >= 'a'))
	}
	return string(out)
}

func b2i(b bool) byte {
	if b {
		return 1
	}
	return 0
}
