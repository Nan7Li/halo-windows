# Halo Desktop

> **新版客户端在 [`desktop/`](desktop/README.md)**：用 MyGo 写的 Go 原生客户端，同一套代码支持 Windows、macOS 和 Linux。下面的 WPF 版本（`src/`）只支持 Windows，保留供参考，之后不再更新。

Windows 原生客户端，给 [Halo](https://github.com/Nan7Li/VoCat)（基于 VoCat 的个人发行版）用。

这不是网页套壳：界面是 **WPF** 自己画的，数据走 Halo 的 HTTP API（Cookie + CSRF）。布局接近桌面聊天软件（左侧栏 + 会话），视觉走 iOS 26/27 那一路的玻璃、圆角和过冲动画。

## 能做什么

- 连到 `http://主机:7575` 的 Halo 实例并登录
- **今天**：主机负载、在线模组、接下来的任务
- **消息**：短信会话列表 + 气泡对话 + 发送
- **电话**：拨号盘、最近通话、回拨（WiFi Calling / IMS）
- **模组**：信号、制式、飞行模式、WiFi Calling、重启
- **链路**：WireGuard 隧道状态
- **现场**：日志
- 没有实例时可以用 **本地预览** 先看界面

模组控制、IMS、短信、eSIM 等能力来自上游 [VoCat](https://github.com/MengMengCode/VoCat)。本仓库只做桌面壳和 API 客户端。

## 要求

- Windows 10 1809+ / Windows 11
- [.NET 8 SDK](https://dotnet.microsoft.com/download/dotnet/8.0)（编译）
- 运行时会随 `dotnet publish` 自带，或安装 .NET 8 Desktop Runtime

## 编译

```powershell
cd $env:USERPROFILE\halo-windows
dotnet publish src\Halo.Desktop\Halo.Desktop.csproj -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true -o dist
```

运行 `dist\Halo.exe`。

登录框填 Halo 地址、管理员用户名和密码。Halo 默认端口 `7575`。

## 许可

Halo / VoCat 本体遵循 Vocat Research & Evaluation License。本客户端代码按仓库内 LICENSE 文件执行；若未单列，与 Halo 分支相同。
