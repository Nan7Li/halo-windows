using System.Runtime.InteropServices;
using System.Windows;
using System.Windows.Interop;
using System.Windows.Media;

namespace Halo.Desktop.Native;

/// <summary>
/// DWM integration: system rounded corners, immersive dark title/frame, and the Mica
/// backdrop (the Windows equivalent of the macOS window material).
///
/// Mica needs a non-layered window, so MainWindow uses WindowChrome instead of
/// AllowsTransparency, extends the DWM frame over the whole client area and clears the
/// WPF composition background. On Windows 10 (no Mica) the window falls back to the
/// opaque "grouped" background colour.
/// </summary>
internal static class WindowFx
{
    const int DWMWA_USE_IMMERSIVE_DARK_MODE = 20;
    const int DWMWA_WINDOW_CORNER_PREFERENCE = 33;
    const int DWMWA_SYSTEMBACKDROP_TYPE = 38;   // Windows 11 22H2+
    const int DWMWA_MICA_EFFECT = 1029;         // Windows 11 21H2 (undocumented)
    const int DWMWCP_ROUND = 2;
    const int DWMSBT_MAINWINDOW = 2;            // Mica

    [DllImport("dwmapi.dll")]
    static extern int DwmSetWindowAttribute(IntPtr hwnd, int attr, ref int value, int size);

    [DllImport("dwmapi.dll")]
    static extern int DwmExtendFrameIntoClientArea(IntPtr hwnd, ref Margins margins);

    [StructLayout(LayoutKind.Sequential)]
    struct Margins { public int Left, Right, Top, Bottom; }

    const int GWL_STYLE = -16;
    const long WS_SYSMENU = 0x00080000;

    [DllImport("user32.dll", EntryPoint = "GetWindowLongPtrW")]
    static extern IntPtr GetWindowLongPtr(IntPtr hwnd, int index);

    [DllImport("user32.dll", EntryPoint = "SetWindowLongPtrW")]
    static extern IntPtr SetWindowLongPtr(IntPtr hwnd, int index, IntPtr value);

    static int Build => Environment.OSVersion.Version.Build;

    /// <summary>True when the OS can draw Mica (Windows 11, build 22000+).</summary>
    public static bool MicaSupported => Environment.OSVersion.Version.Major >= 10 && Build >= 22000;

    /// <summary>Set once the backdrop was actually accepted by DWM for this process.</summary>
    public static bool MicaActive { get; private set; }

    public static void Attach(Window window, Func<bool> isDark)
    {
        window.SourceInitialized += (_, _) =>
        {
            var hwnd = new WindowInteropHelper(window).Handle;
            if (hwnd == IntPtr.Zero) return;

            int round = DWMWCP_ROUND;
            DwmSetWindowAttribute(hwnd, DWMWA_WINDOW_CORNER_PREFERENCE, ref round, sizeof(int));
            SetDark(hwnd, isDark());

            if (!MicaSupported) { MicaActive = false; return; }
            if (HwndSource.FromHwnd(hwnd) is { CompositionTarget: { } target })
                target.BackgroundColor = Colors.Transparent;
            var m = new Margins { Left = -1, Right = -1, Top = -1, Bottom = -1 };
            DwmExtendFrameIntoClientArea(hwnd, ref m);

            // With the frame extended over the client area DWM would also paint its own
            // min/max/close under ours; without WS_SYSMENU it does not. Alt+F4 still closes.
            var style = GetWindowLongPtr(hwnd, GWL_STYLE).ToInt64();
            SetWindowLongPtr(hwnd, GWL_STYLE, new IntPtr(style & ~WS_SYSMENU));

            int hr;
            if (Build >= 22621)
            {
                int mica = DWMSBT_MAINWINDOW;
                hr = DwmSetWindowAttribute(hwnd, DWMWA_SYSTEMBACKDROP_TYPE, ref mica, sizeof(int));
            }
            else
            {
                int on = 1;
                hr = DwmSetWindowAttribute(hwnd, DWMWA_MICA_EFFECT, ref on, sizeof(int));
            }
            MicaActive = hr == 0;
        };
    }

    public static void SetDark(Window window, bool dark)
    {
        var hwnd = new WindowInteropHelper(window).Handle;
        if (hwnd != IntPtr.Zero) SetDark(hwnd, dark);
    }

    static void SetDark(IntPtr hwnd, bool dark)
    {
        int v = dark ? 1 : 0;
        DwmSetWindowAttribute(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE, ref v, sizeof(int));
    }
}
