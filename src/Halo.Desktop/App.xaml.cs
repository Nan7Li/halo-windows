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
            MessageBox.Show(args.Exception.Message, "Halo", MessageBoxButton.OK, MessageBoxImage.Warning);
            args.Handled = true;
        };
        ThemeManager.Apply(AppState.Current.Prefs);
        base.OnStartup(e);

        var i = Array.IndexOf(e.Args, "--shots");
        if (i >= 0 && i + 1 < e.Args.Length)
        {
            var dir = e.Args[i + 1];
            Activated += Once;
            async void Once(object? s, EventArgs _)
            {
                Activated -= Once;
                if (MainWindow is MainWindow w) await Dev.Shots.Run(w, dir);
            }
        }
    }
}
