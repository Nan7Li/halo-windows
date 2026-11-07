# Captures the real window (including the DWM Mica backdrop) for each page in preview mode via UI Automation.
# Usage: powershell -File tools\shoot-window.ps1 -Exe <Halo.exe> -Out <dir> [-Preview <button name>] [-Nav <comma list>] [-Names <comma list>]
param([string]$Exe, [string]$Out, [string]$Preview = "本地预览（不连接实例）",
      [string]$Nav = "主页,消息,电话,设备,网络,日志,设置",
      [string]$Names = "home,messages,phone,devices,network,logs,settings",
      [string]$Appearance = "")
Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes, System.Drawing, System.Windows.Forms
Add-Type @"
using System; using System.Runtime.InteropServices;
public static class W { [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out R r);
 [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
 public struct R { public int L, T, Rr, B; } }
"@
New-Item -ItemType Directory -Force $Out | Out-Null
if ($Appearance) { $p = Start-Process $Exe -ArgumentList "--appearance", $Appearance -PassThru } else { $p = Start-Process $Exe -PassThru }
$AE = [System.Windows.Automation.AutomationElement]
$win = $null
for ($i = 0; $i -lt 40 -and -not $win; $i++) {
  Start-Sleep -Milliseconds 250; $p.Refresh()
  if ($p.MainWindowHandle -ne 0) { $win = $AE::FromHandle($p.MainWindowHandle) }
}
function Find($name, $type = $null) {
  $c = New-Object System.Windows.Automation.PropertyCondition($AE::NameProperty, $name)
  if ($type) {
    $t = New-Object System.Windows.Automation.PropertyCondition($AE::ControlTypeProperty, $type)
    $c = New-Object System.Windows.Automation.AndCondition($c, $t)
  }
  $win.FindFirst([System.Windows.Automation.TreeScope]::Descendants, $c)
}
function Dismiss {
  # The pre-redesign build pops a MessageBox for a TwoWay-binding bug; close them so the page is visible.
  for ($k = 0; $k -lt 30; $k++) {
    $c = New-Object System.Windows.Automation.PropertyCondition($AE::ProcessIdProperty, $p.Id)
    $tops = $AE::RootElement.FindAll([System.Windows.Automation.TreeScope]::Children, $c)
    $dlg = $tops | Where-Object { $_.Current.ClassName -eq '#32770' } | Select-Object -First 1
    if (-not $dlg) { break }
    $dlg.GetCurrentPattern([System.Windows.Automation.WindowPattern]::Pattern).Close()
    Start-Sleep -Milliseconds 150
  }
}
function Shot($file) {
  Start-Sleep -Milliseconds 700
  Dismiss
  Start-Sleep -Milliseconds 300
  # Screen capture copies whatever is on top, so never save a frame unless Halo is in front:
  # otherwise another app's window (and whatever is in it) would end up in the image.
  $front = $false
  for ($t = 0; $t -lt 10 -and -not $front; $t++) {
    [W]::SetForegroundWindow($p.MainWindowHandle) | Out-Null
    Start-Sleep -Milliseconds 200
    $front = [W]::GetForegroundWindow() -eq $p.MainWindowHandle
  }
  if (-not $front) { Write-Warning "Halo is not the foreground window; skipped $file"; return }
  $r = New-Object W+R; [W]::GetWindowRect($p.MainWindowHandle, [ref]$r) | Out-Null
  $bmp = New-Object System.Drawing.Bitmap ($r.Rr - $r.L), ($r.B - $r.T)
  $g = [System.Drawing.Graphics]::FromImage($bmp)
  $g.CopyFromScreen($r.L, $r.T, 0, 0, $bmp.Size); $bmp.Save((Join-Path $Out $file)); $g.Dispose(); $bmp.Dispose()
}
Shot "00-login.png"
(Find $Preview).GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern).Invoke()
$n = 1
$navs = $Nav.Split(","); $files = $Names.Split(",")
for ($k = 0; $k -lt $navs.Length; $k++) {
  $e = Find $navs[$k] ([System.Windows.Automation.ControlType]::RadioButton)
  if ($e) { $e.GetCurrentPattern([System.Windows.Automation.SelectionItemPattern]::Pattern).Select() }
  Shot ("{0:00}-{1}.png" -f $n, $files[$k]); $n++
}
$p.Kill()
