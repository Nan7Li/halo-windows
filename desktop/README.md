# Halo Desktop（MyGo）

Halo 的桌面客户端，一套 Go 代码同时构建 **Windows、macOS 和 Linux** 版本。

界面不是网页套壳。它用 [MyGo](https://github.com/egoist/mygo) 的原生 UI 在 GPU 上自己绘制（Windows 用 Direct3D 11，macOS 用 Metal，Linux 用 OpenGL），侧栏和通话横幅用 MyGo 的 Liquid Glass 插件。整个程序是一个 13–15 MB 的可执行文件，启动快，占内存少，不依赖 WebView。

## 功能

- **登录**：输入 Halo 地址（默认端口 7575）、用户名和密码。密码保存在系统凭据库（macOS 钥匙串、Windows 凭据管理器、Linux Secret Service），下次启动自动登录。
- **HTTPS 自签名证书**：第一次连接时显示证书的 SHA-256 指纹，确认后只信任这一张证书（证书固定），不会关闭证书校验。
- **概览**：主机负载、各条线路的信号和状态、最近的短信和通话、自动任务。
- **信息**：会话列表、搜索、气泡对话、按回车发送（Shift+回车换行）、新建短信并选择发送线路。
- **电话**：拨号键盘（支持键盘直接输入数字），可选择线路；最近通话可筛选全部或未接，一键回拨；来电时窗口顶部出现横幅，可接听或挂断。
- **设备**：信号、网络制式、频段、公网 IP、IMEI/ICCID；可开关飞行模式和 WiFi 通话、重连 WiFi 通话、重启模组。
- **网络**：WireGuard 隧道的连接和断开。
- **日志**：按级别筛选、搜索，自动滚动到最新一条。
- **设置**：浅色、深色或跟随系统；强调色；侧栏玻璃的通透度；新短信和来电的系统通知。
- **预览模式**：登录页选「不登录，先看看界面」，用示例数据浏览所有页面。

### 暂不支持

- **通话音频**：MyGo 目前只能播放声音，不能采集麦克风，所以客户端只负责控制通话（拨号、接听、挂断）。要通话请用浏览器打开 Halo 网页版。
- 网页版的高级管理功能（eSIM、AT 指令、代理、自动任务编辑等）仍在网页版里。设置页有「在浏览器中打开 Halo」。

## 下载

在 GitHub Actions 的 **Desktop** 工作流里下载构建产物；推送 `desktop-v*` 标签后，产物会附在草稿 Release 里：

| 系统 | 文件 |
|---|---|
| Windows 10/11（x64、ARM64） | `windows-*-halo-desktop-setup-*.exe`（安装包），或 `Halo.exe` 直接运行 |
| macOS 12+（Intel 和 Apple 芯片） | `*.dmg` 或 `Halo-macOS.app.zip` |
| Linux（x64、ARM64） | `.deb` 或 `.tar.gz`（内含 `install.sh`） |

**macOS 版没有用 Apple 开发者证书签名**，首次打开时在 Finder 里右键点击 Halo，选择「打开」；如果提示「已损坏」，运行 `xattr -dr com.apple.quarantine /Applications/Halo.app`。我们手上没有 Mac，macOS 版只做了编译和 CI 构建，没有真机测试，有问题请提 issue。

Linux 需要 GTK 3 和 OpenGL 驱动，`.deb` 会自动安装依赖。

## 开发

需要 Go 1.27.3+，不需要 cgo，在任何系统上都能为三个平台交叉编译。

```sh
cd desktop
go test ./...                     # 不开窗口跑界面测试（模拟点击、输入，验证每个页面）
go run .                          # 运行
go tool mygo dev                  # 改代码后自动重载
go tool mygo build -platform windows/amd64,linux/amd64   # 打包；macOS 的 dmg 需在 macOS 上构建
```

`SNAPSHOT_DIR=/某个目录 go test ./...` 会把每个页面的浅色和深色截图写成 PNG。`HALO_WRITE_ICON=1 go test -run TestWriteIcon` 会重新生成 `resources/icon.png`。

| 文件 | 内容 |
|---|---|
| `api.go` | Halo HTTP API 客户端：会话 Cookie、双提交 CSRF、证书固定 |
| `model.go` | 数据结构和格式化 |
| `app.go` | 应用状态、登录、轮询、通知、操作 |
| `shell.go` | 窗口框架：玻璃侧栏、标题栏、来电横幅、通用组件 |
| `page_*.go` | 各页面 |
| `theme.go` | iOS 系统配色和强调色 |
| `demo.go` | 预览模式的示例数据 |
