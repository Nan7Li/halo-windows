using System.IO;
using System.Windows;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Threading;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Dev;

/// <summary>
/// Screenshot harness: <c>Halo.exe --shots &lt;dir&gt;</c>.
/// Enters preview mode and renders every page in light and dark at 96 and 144 DPI
/// (100 % / 150 % scaling) with motion off, plus a light 100 % set at the 960 px minimum
/// width. WPF lays out in device-independent units, so a 144 DPI render is the same
/// layout a 150 % monitor gets. DWM Mica is not part of the WPF visual tree, so it is
/// replaced with the opaque fallback colour in these images. Preferences are not saved.
/// </summary>
internal static class Shots
{
    static readonly (HaloPage page, string name)[] Pages =
    [
        (HaloPage.Home, "home"), (HaloPage.Messages, "messages"), (HaloPage.Phone, "phone"),
        (HaloPage.Devices, "devices"), (HaloPage.Network, "network"), (HaloPage.Logs, "logs"),
        (HaloPage.Settings, "settings"),
    ];

    /// <summary>
    /// <c>--anim</c>: samples the IosSwitch thumb and SegmentedControl indicator every frame for
    /// each motion level and writes anim.txt, so the spring/overshoot and the Reduced/Off
    /// behaviour can be checked numerically.
    /// </summary>
    public static async Task Anim(MainWindow w, string dir)
    {
        Directory.CreateDirectory(dir);
        var prefs = AppState.Current.Prefs;
        await w.EnterPreviewAsync();
        var log = new System.Text.StringBuilder();
        foreach (var level in new[] { "full", "reduced", "off" })
        {
            prefs.Motion = level;
            ThemeManager.Apply(prefs);

            w.ShowPage(HaloPage.Settings, animate: false);
            await Settle(w);
            var sw = (Controls.IosSwitch)w.SetPage.FindName("BoldSwitch");
            var thumb = (FrameworkElement)sw.Template.FindName("PART_Thumb", sw);
            var on = (FrameworkElement)sw.Template.FindName("PART_On", sw);
            sw.IsChecked = false;
            await Settle(w);
            log.AppendLine($"[{level}] IosSwitch off->on (travel 21): t_ms thumbX greenOpacity");
            sw.IsChecked = true;
            await Sample(w, log, () => $"{((TranslateTransform)thumb.RenderTransform).X:0.00} {on.Opacity:0.00}");
            sw.IsChecked = false;

            w.ShowPage(HaloPage.Logs, animate: false);
            await Settle(w);
            var seg = (Controls.SegmentedControl)w.LogsPage.FindName("LevelPick");
            var ind = (FrameworkElement)seg.Template.FindName("PART_Indicator", seg);
            seg.SelectedIndex = 0;
            await Settle(w);
            var segW = (seg.ActualWidth - 4) / seg.Items.Count;
            log.AppendLine($"[{level}] Segmented 0->2 (target {2 * segW:0.00}): t_ms indicatorX opacity");
            seg.SelectedIndex = 2;
            await Sample(w, log, () => $"{((TranslateTransform)ind.RenderTransform).X:0.00} {ind.Opacity:0.00}");
            seg.SelectedIndex = 0;
        }
        File.WriteAllText(Path.Combine(dir, "anim.txt"), log.ToString());
        Application.Current.Shutdown();
    }

    static async Task Sample(Window w, System.Text.StringBuilder log, Func<string> read)
    {
        var sw = System.Diagnostics.Stopwatch.StartNew();
        while (sw.ElapsedMilliseconds < 420)
        {
            await w.Dispatcher.InvokeAsync(() => { }, DispatcherPriority.Render);
            log.AppendLine($"  {sw.ElapsedMilliseconds,4} {read()}");
            await Task.Delay(15);
        }
    }

    public static async Task Run(MainWindow w, string dir, bool bold = false)
    {
        Directory.CreateDirectory(dir);
        var prefs = AppState.Current.Prefs;
        prefs.Motion = "off";
        prefs.BoldText = bold;
        w.WindowState = WindowState.Normal;
        w.Width = 1180; w.Height = 760;
        await Settle(w);

        foreach (var theme in new[] { "light", "dark" })
        {
            prefs.Appearance = theme;
            ThemeManager.Apply(prefs);
            await Settle(w);
            foreach (var scale in new[] { 1.0, 1.5 })
                Save(w, Path.Combine(dir, $"{theme}-{scale * 100:0}-00-login.png"), scale);
        }

        await w.EnterPreviewAsync();
        w.MsgPage.SelectFirst();
        foreach (var theme in new[] { "light", "dark" })
        {
            prefs.Appearance = theme;
            ThemeManager.Apply(prefs);
            for (var i = 0; i < Pages.Length; i++)
            {
                w.ShowPage(Pages[i].page, animate: false);
                await Settle(w);
                foreach (var scale in new[] { 1.0, 1.5 })
                    Save(w, Path.Combine(dir, $"{theme}-{scale * 100:0}-{i + 1:00}-{Pages[i].name}.png"), scale);
            }
        }

        // Narrowest supported window.
        prefs.Appearance = "light";
        ThemeManager.Apply(prefs);
        w.Width = w.MinWidth; w.Height = 640;
        for (var i = 0; i < Pages.Length; i++)
        {
            w.ShowPage(Pages[i].page, animate: false);
            await Settle(w);
            Save(w, Path.Combine(dir, $"min960-{i + 1:00}-{Pages[i].name}.png"), 1.0);
        }
        Application.Current.Shutdown();
    }

    static async Task Settle(Window w)
    {
        w.UpdateLayout();
        await w.Dispatcher.InvokeAsync(() => { }, DispatcherPriority.ApplicationIdle);
        await Task.Delay(350);
        w.UpdateLayout();
        await w.Dispatcher.InvokeAsync(() => { }, DispatcherPriority.ApplicationIdle);
    }

    static void Save(MainWindow w, string file, double scale)
    {
        var root = (FrameworkElement)w.Content;
        Controls.Hairline.ScaleOverride = scale;
        Invalidate(root);
        root.UpdateLayout();
        var size = new Size(root.ActualWidth, root.ActualHeight);
        var dv = new DrawingVisual();
        using (var dc = dv.RenderOpen())
        {
            dc.DrawRectangle((Brush)Application.Current.FindResource("GroupedBrush"), null, new Rect(size));
            dc.DrawRectangle(new VisualBrush(root) { Stretch = Stretch.None, AlignmentX = AlignmentX.Left, AlignmentY = AlignmentY.Top },
                null, new Rect(size));
        }
        var bmp = new RenderTargetBitmap((int)Math.Ceiling(size.Width * scale), (int)Math.Ceiling(size.Height * scale),
            96 * scale, 96 * scale, PixelFormats.Pbgra32);
        bmp.Render(dv);
        var enc = new PngBitmapEncoder();
        enc.Frames.Add(BitmapFrame.Create(bmp));
        using var fs = File.Create(file);
        enc.Save(fs);
    }

    static void Invalidate(DependencyObject d)
    {
        if (d is Controls.Hairline h) { h.InvalidateMeasure(); h.InvalidateVisual(); }
        for (var i = 0; i < VisualTreeHelper.GetChildrenCount(d); i++) Invalidate(VisualTreeHelper.GetChild(d, i));
    }
}
