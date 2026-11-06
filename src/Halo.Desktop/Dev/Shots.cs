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

    public static async Task Run(MainWindow w, string dir)
    {
        Directory.CreateDirectory(dir);
        var prefs = AppState.Current.Prefs;
        prefs.Motion = "off";
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
