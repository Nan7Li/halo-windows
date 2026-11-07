using System.Windows;
using System.Windows.Media;
using Microsoft.Win32;

namespace Halo.Desktop.Services;

/// <summary>
/// Applies appearance preferences at runtime: light/dark palette (or follow Windows),
/// accent colour, glass clarity, motion level and Bold Text. All styles consume these
/// keys through DynamicResource, so switching is live.
/// </summary>
public static class ThemeManager
{
    public sealed record AccentOption(string Key, string Name, Color Light, Color Dark);

    /// <summary>macOS accent palette (light / dark variants).</summary>
    public static readonly AccentOption[] Accents =
    [
        new("blue", "蓝色", Color.FromRgb(0, 122, 255), Color.FromRgb(10, 132, 255)),
        new("purple", "紫色", Color.FromRgb(175, 82, 222), Color.FromRgb(191, 90, 242)),
        new("pink", "粉色", Color.FromRgb(255, 45, 85), Color.FromRgb(255, 55, 95)),
        new("red", "红色", Color.FromRgb(255, 59, 48), Color.FromRgb(255, 69, 58)),
        new("orange", "橙色", Color.FromRgb(255, 149, 0), Color.FromRgb(255, 159, 10)),
        new("yellow", "黄色", Color.FromRgb(255, 204, 0), Color.FromRgb(255, 214, 10)),
        new("green", "绿色", Color.FromRgb(52, 199, 89), Color.FromRgb(48, 209, 88)),
        new("graphite", "石墨色", Color.FromRgb(142, 142, 147), Color.FromRgb(152, 152, 157)),
    ];

    public static bool IsDark { get; private set; }
    public static event Action? Changed;

    static readonly Uri LightUri = new("pack://application:,,,/Halo;component/Themes/Light.xaml");
    static readonly Uri DarkUri = new("pack://application:,,,/Halo;component/Themes/Dark.xaml");
    static ResourceDictionary? _light, _dark;

    public static bool SystemIsDark()
    {
        try
        {
            using var key = Registry.CurrentUser.OpenSubKey(@"Software\Microsoft\Windows\CurrentVersion\Themes\Personalize");
            return key?.GetValue("AppsUseLightTheme") is int v && v == 0;
        }
        catch { return false; }
    }

    public static MotionLevel ParseMotion(string? s) => s switch
    {
        "reduced" => MotionLevel.Reduced,
        "off" => MotionLevel.Off,
        _ => MotionLevel.Full
    };

    public static void Apply(Prefs p)
    {
        var app = Application.Current;
        if (app == null) return;
        var res = app.Resources;
        var dark = p.Appearance switch { "light" => false, "dark" => true, _ => SystemIsDark() };
        IsDark = dark;

        // 1. Palette: index 0 of the merged dictionaries is always the palette.
        _light ??= new ResourceDictionary { Source = LightUri };
        _dark ??= new ResourceDictionary { Source = DarkUri };
        var palette = dark ? _dark : _light;
        var merged = res.MergedDictionaries;
        if (merged.Count == 0) merged.Add(palette);
        else if (!ReferenceEquals(merged[0], palette)) merged[0] = palette;

        // 2. Accent.
        var accent = Accents.FirstOrDefault(a => a.Key == p.Accent) ?? Accents[0];
        var ac = dark ? accent.Dark : accent.Light;
        res["AccentColor"] = ac;
        res["AccentBrush"] = Frozen(ac);
        res["AccentTintBrush"] = Frozen(Color.FromArgb(38, ac.R, ac.G, ac.B));

        // 3. Glass. WPF cannot blur what is behind an element, so the blur comes from the
        //    DWM Mica backdrop under the sidebar; the slider sets how much tint sits on top.
        var t = Math.Clamp(p.Glass, 0, 1);
        var baseC = dark ? Color.FromRgb(36, 36, 38) : Color.FromRgb(255, 255, 255);
        res["GlassFillBrush"] = Frozen(Color.FromArgb((byte)(255 * Lerp(0.30, 0.86, t)), baseC.R, baseC.G, baseC.B));
        res["MenuFillBrush"] = Frozen(Color.FromArgb((byte)(255 * Lerp(0.88, 0.97, t)),
            dark ? (byte)40 : (byte)248, dark ? (byte)40 : (byte)248, dark ? (byte)42 : (byte)250));
        res["GlassStrokeBrush"] = Rim(dark);

        // 4. Bold Text: every weight steps up one notch.
        res["W.Regular"] = p.BoldText ? FontWeights.Medium : FontWeights.Normal;
        res["W.Medium"] = p.BoldText ? FontWeights.SemiBold : FontWeights.Medium;
        res["W.Semibold"] = p.BoldText ? FontWeights.Bold : FontWeights.SemiBold;

        // 5. Motion.
        Motion.Level = ParseMotion(p.Motion);

        Changed?.Invoke();
    }

    /// <summary>1px rim: bright at the top edge, fading out, slightly dark at the bottom.</summary>
    static Brush Rim(bool dark)
    {
        var b = new LinearGradientBrush { StartPoint = new Point(0, 0), EndPoint = new Point(0, 1) };
        if (dark)
        {
            b.GradientStops.Add(new GradientStop(Color.FromArgb(90, 255, 255, 255), 0));
            b.GradientStops.Add(new GradientStop(Color.FromArgb(28, 255, 255, 255), 0.35));
            b.GradientStops.Add(new GradientStop(Color.FromArgb(16, 255, 255, 255), 1));
        }
        else
        {
            b.GradientStops.Add(new GradientStop(Color.FromArgb(255, 255, 255, 255), 0));
            b.GradientStops.Add(new GradientStop(Color.FromArgb(110, 255, 255, 255), 0.35));
            b.GradientStops.Add(new GradientStop(Color.FromArgb(22, 0, 0, 0), 1));
        }
        b.Freeze();
        return b;
    }

    static double Lerp(double a, double b, double t) => a + (b - a) * t;

    static SolidColorBrush Frozen(Color c)
    {
        var b = new SolidColorBrush(c);
        b.Freeze();
        return b;
    }
}
