using System.IO;
using System.Windows;
using System.Windows.Interop;
using System.Windows.Media;
using System.Windows.Threading;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Dev;

/// <summary>
/// Frame-time probe: <c>Halo.exe --perf &lt;dir&gt; [--max]</c>.
/// Runs the real window (motion = full) through page switches, sidebar moves, switch and
/// segmented animations and a data refresh, records the interval between rendered frames
/// (CompositionTarget.Rendering) and writes perf.txt. Preferences are not saved.
/// </summary>
internal static class Perf
{
    sealed class Probe
    {
        // Wall-clock gaps between rendered frames, counted from Start(), so the frame that
        // waits for synchronous work done right after Start() (layout, binding) is included.
        readonly List<double> _ms = [];
        readonly System.Diagnostics.Stopwatch _clock = new();
        TimeSpan _lastRender, _last;
        public void Start() { _ms.Clear(); _lastRender = TimeSpan.Zero; _clock.Restart(); _last = TimeSpan.Zero; CompositionTarget.Rendering += Tick; }
        public void Stop() => CompositionTarget.Rendering -= Tick;
        void Tick(object? s, EventArgs e)
        {
            var t = ((RenderingEventArgs)e).RenderingTime;
            if (t == _lastRender) return;
            _lastRender = t;
            var now = _clock.Elapsed;
            _ms.Add((now - _last).TotalMilliseconds);
            _last = now;
        }
        public string Summary()
        {
            if (_ms.Count == 0) return "no frames";
            var sorted = _ms.OrderBy(x => x).ToList();
            var p95 = sorted[(int)Math.Min(sorted.Count - 1, Math.Ceiling(sorted.Count * 0.95) - 1)];
            var slow = _ms.Count(x => x > 25);
            return $"frames={_ms.Count,3} avg={_ms.Average(),5:0.0}ms p95={p95,5:0.0}ms max={_ms.Max(),6:0.0}ms dropped(>25ms)={slow}";
        }
        public List<double> Raw => _ms;
    }

    public static async Task Run(MainWindow w, string dir, bool maximize)
    {
        Directory.CreateDirectory(dir);
        var log = new System.Text.StringBuilder();
        var prefs = AppState.Current.Prefs;
        prefs.Motion = "full";
        ThemeManager.Apply(prefs);
        if (maximize) w.WindowState = WindowState.Maximized;
        await Idle(w, 200);

        var dpi = VisualTreeHelper.GetDpi(w);
        log.AppendLine($"RenderCapability.Tier={RenderCapability.Tier >> 16}  dpi={dpi.PixelsPerDip * 100:0}%  " +
                       $"window={w.ActualWidth:0}x{w.ActualHeight:0}  mica={Native.WindowFx.MicaActive}  " +
                       $"renderMode={(HwndSource.FromVisual(w) as HwndSource)?.CompositionTarget.RenderMode}  os={Environment.OSVersion.Version}");

        await w.EnterPreviewAsync();
        await Idle(w, 800);
        var probe = new Probe();

        // 1. Page switches: cross-fade + sidebar capsule spring.
        var order = new[] { HaloPage.Messages, HaloPage.Phone, HaloPage.Devices, HaloPage.Network, HaloPage.Logs, HaloPage.Settings, HaloPage.Home };
        var all = new List<double>();
        for (var round = 1; round <= 2; round++)
        {
        all.Clear();
        foreach (var page in order)
        {
            probe.Start();
            var sw0 = System.Diagnostics.Stopwatch.StartNew();
            w.ShowPage(page);
            var showMs = sw0.Elapsed.TotalMilliseconds;
            w.UpdateLayout();
            var layoutMs = sw0.Elapsed.TotalMilliseconds - showMs;
            await Task.Delay(450);
            probe.Stop();
            log.AppendLine($"r{round} switch → {page,-9} {probe.Summary()}  [ShowPage {showMs:0.0}ms, layout {layoutMs:0.0}ms]");
            all.AddRange(probe.Raw);
            await Idle(w, 300);
        }
        log.AppendLine($"r{round} switch ALL      {Summ(all)}");
        // The 4 s poll runs between rounds, as it would while the user clicks around.
        await w.RefreshAll();
        await Idle(w, 300);
        }

        // 2. IosSwitch on Settings.
        w.ShowPage(HaloPage.Settings, animate: false);
        await Idle(w, 400);
        var sw = (Controls.IosSwitch)w.SetPage.FindName("BoldSwitch");
        all.Clear();
        for (var i = 0; i < 4; i++)
        {
            probe.Start();
            sw.IsChecked = !(sw.IsChecked == true);
            await Task.Delay(420);
            probe.Stop();
            all.AddRange(probe.Raw);
        }
        sw.IsChecked = false;
        log.AppendLine($"IosSwitch x4       {Summ(all)}");

        // 3. Segmented control on Logs.
        w.ShowPage(HaloPage.Logs, animate: false);
        await Idle(w, 400);
        var seg = (Controls.SegmentedControl)w.LogsPage.FindName("LevelPick");
        all.Clear();
        foreach (var idx in new[] { 3, 0, 2, 0 })
        {
            probe.Start();
            seg.SelectedIndex = idx;
            await Task.Delay(420);
            probe.Stop();
            all.AddRange(probe.Raw);
        }
        log.AppendLine($"Segmented x4       {Summ(all)}");

        // 4. Background refresh (what the 4 s poll does) while a page switch is animating.
        w.ShowPage(HaloPage.Home, animate: false);
        await Idle(w, 400);
        probe.Start();
        w.ShowPage(HaloPage.Devices);
        var t0 = System.Diagnostics.Stopwatch.StartNew();
        await w.RefreshAll();
        var refreshMs = t0.ElapsedMilliseconds;
        await Task.Delay(450);
        probe.Stop();
        log.AppendLine($"switch+refresh     {probe.Summary()}  (RefreshAll took {refreshMs} ms)");

        File.WriteAllText(Path.Combine(dir, "perf.txt"), log.ToString());
        Application.Current.Shutdown();
    }

    static string Summ(List<double> ms)
    {
        if (ms.Count == 0) return "no frames";
        var sorted = ms.OrderBy(x => x).ToList();
        var p95 = sorted[(int)Math.Min(sorted.Count - 1, Math.Ceiling(sorted.Count * 0.95) - 1)];
        return $"frames={ms.Count,3} avg={ms.Average(),5:0.0}ms p95={p95,5:0.0}ms max={ms.Max(),6:0.0}ms dropped(>25ms)={ms.Count(x => x > 25)}";
    }

    static async Task Idle(Window w, int ms)
    {
        await w.Dispatcher.InvokeAsync(() => { }, DispatcherPriority.ApplicationIdle);
        await Task.Delay(ms);
    }
}
