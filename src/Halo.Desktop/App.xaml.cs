using System.Windows;
using System.Windows.Media;
using Halo.Desktop.Services;

namespace Halo.Desktop;

public partial class App : Application
{
    protected override void OnStartup(StartupEventArgs e)
    {
        RenderOptions.ProcessRenderMode = System.Windows.Interop.RenderMode.Default;
        DispatcherUnhandledException += (_, args) =>
        {
            try
            {
                System.IO.File.AppendAllText(System.IO.Path.Combine(System.IO.Path.GetTempPath(), "halo-desktop-errors.log"),
                    $"{DateTime.Now:O}{Environment.NewLine}{args.Exception}{Environment.NewLine}{Environment.NewLine}");
            }
            catch { /* logging is best-effort */ }
            MessageBox.Show(args.Exception.Message, "Halo", MessageBoxButton.OK, MessageBoxImage.Warning);
            args.Handled = true;
        };
        // --appearance light|dark overrides the saved preference for this run only (screenshots).
        var a = Array.IndexOf(e.Args, "--appearance");
        if (a >= 0 && a + 1 < e.Args.Length) AppState.Current.Prefs.Appearance = e.Args[a + 1];
        ThemeManager.Apply(AppState.Current.Prefs);
        base.OnStartup(e);

        var pi = Array.IndexOf(e.Args, "--perf");
        if (pi >= 0 && pi + 1 < e.Args.Length)
        {
            var dir = e.Args[pi + 1];
            Activated += OncePerf;
            async void OncePerf(object? s, EventArgs _)
            {
                Activated -= OncePerf;
                if (MainWindow is MainWindow w) await Dev.Perf.Run(w, dir, e.Args.Contains("--max"));
            }
            return;
        }

        var i = Array.IndexOf(e.Args, "--shots");
        if (i >= 0 && i + 1 < e.Args.Length)
        {
            var dir = e.Args[i + 1];
            Activated += Once;
            async void Once(object? s, EventArgs _)
            {
                Activated -= Once;
                if (MainWindow is not MainWindow w) return;
                if (e.Args.Contains("--anim")) await Dev.Shots.Anim(w, dir);
                else await Dev.Shots.Run(w, dir, bold: e.Args.Contains("--bold"));
            }
        }
    }
}
