# Halo Desktop

Windows 原生客户端，给 [Halo](https://github.com/Nan7Li/VoCat)（基于 VoCat 的个人发行版）用。

这不是网页套壳：界面是 **WPF** 自己画的，数据走 Halo 的 HTTP API（Cookie + CSRF）。外观和 Halo 网页版的新设计一致：桌面端对齐 macOS（Finder、系统设置、信息 App），导航层用 Liquid Glass 材质。

## 能做什么

- 连到 `http://主机:7575` 的 Halo 实例并登录
- **主页**：线路在线数、即将执行的任务、主机 CPU / 内存 / 磁盘、网络速率、最近通话、最近信息
- **消息**：会话列表（搜索、未读圆点）+ 气泡对话 + 发送
- **电话**：最近通话（未接标红，点按回拨）+ 拨号盘（WiFi Calling / IMS）
- **设备**：模组列表与详情，VoWiFi、飞行模式、重启，IMEI / ICCID 默认模糊
- **网络**：WireGuard 隧道状态
- **日志**：搜索 + 级别筛选
- **设置**：外观、连接、关于
- 没有实例时可以用 **本地预览** 先看界面

模组控制、IMS、短信、eSIM 等能力来自上游 [VoCat](https://github.com/MengMengCode/VoCat)。本仓库只做桌面壳和 API 客户端。

## 视觉

界面参照 Apple Human Interface Guidelines（macOS），在 Windows 上用系统能力实现：

- **颜色**：Apple 系统语义色（label / label2 / label3、grouped / grouped2 / grouped3、separator、fill3，以及 blue / green / orange / red），浅色和深色各一套（`Themes/Light.xaml`、`Themes/Dark.xaml`），默认跟随 Windows，运行时切换。强调色可选，默认系统蓝；状态色只用绿、橙、红。
- **字体**：Segoe UI Variable Text，中文回退微软雅黑。正文 13，说明 11（最小），工具栏标题 15 Semibold 靠左；分组标题 11、label2、不加粗。设置里的「粗体文本」会把全局字重各提高一级。
- **形状**：连续曲率圆角（squircle）。菜单项 6，输入框 / 按钮为胶囊或 10，小卡片 10，分组列表 14，弹窗 18；嵌套圆角保持同心（外 = 内 + 内边距）。主窗口用 DWM 系统圆角。
- **材质**：Windows 11 上窗口背景是 Mica（macOS 窗口材质的对应物），Windows 10 退回纯色 grouped 背景。玻璃只用在导航层和控件层：浮动侧栏（220 宽、距窗口边 8）、工具栏按钮组、菜单、登录卡片；半透明填充加 1px 上亮下暗的高光边。内容区始终是不透明的分组列表。设置里的「玻璃质感」滑块调节玻璃从透明到着色。
- **控件**：iOS 26 开关（63×28，滑块 38×24）、带滑动指示的分段控件、38 高的列表行（分隔线从文字起点开始，一个物理像素）、行内输入的表单、玻璃上下文菜单（项高 24）、无框空状态、15% 底色的提示条。可复用控件在 `Controls/`，样式在 `Themes/`。
- **动效**：页面切换交叉淡化 180 ms；开关、分段滑块、侧栏选中态用弹簧（约 4% 过冲，320 ms）。「动态效果」设为「减弱」时去掉位移和弹跳只留淡入淡出，「关闭」时没有动画。
- **清晰度**：PerMonitorV2 DPI、布局取整、`TextFormattingMode=Display`；图标是 Segoe Fluent Icons（Windows 10 上是同码位的 Segoe MDL2 Assets）字体，100% / 150% / 200% 下都是矢量。

窗口按钮（最小化 / 最大化 / 关闭）保留在右上角，按 Windows 的习惯而不是 macOS 的红绿灯。

改版前后的截图在 `docs/screenshots/`。

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

## 截图工具（开发用）

- `Halo.exe --shots <目录>`：本地预览模式下把每个页面按浅色 / 深色 × 100% / 150% 渲染成 PNG，另加一套 960 宽的最小窗口截图。加 `--bold` 打开粗体文本，`--anim` 改为逐帧记录开关和分段控件的动画数值。这些都不会写入偏好设置。Mica 不在 WPF 渲染树里，所以这里用纯色背景代替。
- `Halo.exe --perf <目录> [--max]`：本地预览模式下自动切换页面、拨动开关和分段控件，记录每帧间隔，写到 `perf.txt`（加 `--max` 先最大化窗口）。帧间隔超过 25 ms 记为掉帧。
- `tools\shoot-window.ps1 -Exe <Halo.exe> -Out <目录> [-Appearance light|dark]`：截真实窗口（含 Mica）。只有 Halo 在最前面时才会保存截图。

## 许可

Halo / VoCat 本体遵循 Vocat Research & Evaluation License。本客户端代码按仓库内 LICENSE 文件执行；若未单列，与 Halo 分支相同。
