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
			card := ui.Column(c).Width(380).Padding(28, 28, 24).Gap(14).Radius(26).Material(glass.Glass{Clarity: 0.2})
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
					in := ui.TextInput(c.Key("server"), &a.prefs.Server).Placeholder("例如 192.168.1.10:7575").Label("服务器地址")
					if a.prefs.Server == "" {
						in.AutoFocus()
					}
					submit = in.Submitted() || submit
				})
				field("用户名", func() {
					submit = ui.TextInput(c.Key("user"), &a.prefs.Username).Placeholder("admin").Label("用户名").Submitted() || submit
				})
				field("密码", func() {
					in := ui.TextInput(c.Key("password"), &a.password).Password().Label("密码")
					if a.prefs.Server != "" && a.prefs.Username != "" {
						in.AutoFocus()
					}
					submit = in.Submitted() || submit
				})
				ui.Checkbox(c, &a.prefs.Remember, "记住密码（保存在系统钥匙串）")
				if a.loginErr != "" {
					ui.Row(c).Gap(6).AlignItems(ui.Start).Children(func() {
						ui.Icon(c, icWarn).FontSize(14).TextColor(p.Red)
						ui.Text(c, a.loginErr).FontSize(12).TextColor(p.Red).Grow(1)
					})
				}
				ready := a.prefs.Server != "" && a.prefs.Username != "" && a.password != ""
				login := ui.PrimaryButton(c, "登录").Disabled(a.busy || !ready).Height(36).Radius(18)
				if a.busy {
					login.Children(func() {
						ui.Spinner(c).Size(14, 14).TextColor(ui.RGB(255, 255, 255))
						ui.Text(c, "正在登录…")
					})
				}
				if (login.Clicked() || submit) && ready && !a.busy {
					server, user, pw := a.prefs.Server, a.prefs.Username, a.password
					a.busy = true
					a.spawn(func() { a.signIn(server, user, pw) })
				}
				if ui.Button(c, "不登录，先看看界面").Height(32).Radius(16).Background(ui.Transparent).Clicked() {
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
