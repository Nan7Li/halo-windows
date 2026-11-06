# Captures each page of the pre-redesign build in preview mode via UI Automation.
# Usage: powershell -File tools\shoot-before.ps1 -Exe <path to Halo.exe> -Out docs\screenshots\before
param([string]$Exe, [string]$Out)
Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes, System.Drawing, System.Windows.Forms
Add-Type @"
using System; using System.Runtime.InteropServices;
public static class W { [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out R r);
 public struct R { public int L, T, Rr, B; } }
"@
New-Item -ItemType Directory -Force $Out | Out-Null
$p = Start-Process $Exe -PassThru
$AE = [System.Windows.Automation.AutomationElement]
$win = $null
for ($i = 0; $i -lt 40 -and -not $win; $i++) {
  Start-Sleep -Milliseconds 250; $p.Refresh()
  if ($p.MainWindowHandle -ne 0) { $win = $AE::FromHandle($p.MainWindowHandle) }
}
function Find($name) {
  $c = New-Object System.Windows.Automation.PropertyCondition($AE::NameProperty, $name)
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
  [W]::SetForegroundWindow($p.MainWindowHandle) | Out-Null
  $r = New-Object W+R; [W]::GetWindowRect($p.MainWindowHandle, [ref]$r) | Out-Null
  $bmp = New-Object System.Drawing.Bitmap ($r.Rr - $r.L), ($r.B - $r.T)
  $g = [System.Drawing.Graphics]::FromImage($bmp)
  $g.CopyFromScreen($r.L, $r.T, 0, 0, $bmp.Size); $bmp.Save((Join-Path $Out $file)); $g.Dispose(); $bmp.Dispose()
}
Shot "00-login.png"
(Find "先看界面（本地预览）").GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern).Invoke()
$n = 1
foreach ($nav in "今天","消息","电话","模组","链路","现场","设置") {
  $e = Find $nav
  if ($e) { $e.GetCurrentPattern([System.Windows.Automation.SelectionItemPattern]::Pattern).Select() }
  Shot ("{0:00}-{1}.png" -f $n, $nav); $n++
}
$p.Kill()
