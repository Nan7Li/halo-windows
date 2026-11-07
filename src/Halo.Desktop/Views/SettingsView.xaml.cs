using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Windows.Threading;
using Halo.Desktop.Native;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class SettingsView : UserControl
{
    static readonly string[] Themes = ["light", "dark", "system"];
    static readonly string[] Motions = ["full", "reduced", "off"];

    readonly Prefs _p = AppState.Current.Prefs;
    readonly DispatcherTimer _save = new() { Interval = TimeSpan.FromMilliseconds(400) };
    AppState? _s;
    bool _init;

    public SettingsView()
    {
        InitializeComponent();
        _save.Tick += (_, _) => { _save.Stop(); _p.Save(); };
        // Keep the controls in sync when prefs change elsewhere (system theme switch, harness).
        ThemeManager.Changed += () => { if (IsLoaded) LoadPrefs(); };
        Loaded += (_, _) => { LoadPrefs(); RenderAbout(); };
        Categories.SelectedIndex = 0;
    }

    public void Bind(AppState s)
    {
        _s = s;
        ConnServer.Text = s.Client.Preview ? "本地预览" : s.Server;
        ConnState.Text = s.Client.Preview ? "未连接实例" : "已连接";
        ConnVersion.Text = string.IsNullOrWhiteSpace(s.Version) ? "—" : s.Version;
        ConnUser.Text = string.IsNullOrWhiteSpace(s.Username) ? "—" : $"{s.Username}（{s.Role}）";
    }

    void LoadPrefs()
    {
        _init = true;
        ThemePick.SelectedIndex = Math.Max(0, Array.IndexOf(Themes, _p.Appearance));
        MotionPick.SelectedIndex = Math.Max(0, Array.IndexOf(Motions, _p.Motion));
        GlassSlider.Value = _p.Glass;
        BoldSwitch.IsChecked = _p.BoldText;
        BuildAccents();
        _init = false;
    }

    void BuildAccents()
    {
        AccentRow.Children.Clear();
        foreach (var a in ThemeManager.Accents)
        {
            var rb = new RadioButton
            {
                Style = (Style)FindResource("AccentDot"),
                Background = new SolidColorBrush(ThemeManager.IsDark ? a.Dark : a.Light),
                ToolTip = a.Name,
                IsChecked = a.Key == _p.Accent,
                Tag = a.Key
            };
            rb.Checked += (_, _) =>
            {
                if (_init || _p.Accent == (string)rb.Tag) return;
                _p.Accent = (string)rb.Tag;
                Commit();
            };
            AccentRow.Children.Add(rb);
        }
    }

    void Commit()
    {
        ThemeManager.Apply(_p);
        _save.Stop();
        _save.Start();
    }

    void CategoryChanged(object sender, SelectionChangedEventArgs e)
    {
        var panes = new FrameworkElement[] { AppearancePane, ConnectionPane, AboutPane };
        var i = Math.Max(0, Categories.SelectedIndex);
        for (var k = 0; k < panes.Length; k++)
        {
            if (k == i) Motion.CrossFade(null, panes[k]);
            else panes[k].Visibility = Visibility.Collapsed;
        }
        if (i == 2) RenderAbout();
    }

    void ThemeChanged(object sender, SelectionChangedEventArgs e)
    {
        if (_init || ThemePick.SelectedIndex < 0) return;
        _p.Appearance = Themes[ThemePick.SelectedIndex];
        Commit();
    }

    void MotionChanged(object sender, SelectionChangedEventArgs e)
    {
        if (_init || MotionPick.SelectedIndex < 0) return;
        _p.Motion = Motions[MotionPick.SelectedIndex];
        Commit();
    }

    void GlassChanged(object sender, RoutedPropertyChangedEventArgs<double> e)
    {
        if (_init) return;
        _p.Glass = Math.Round(e.NewValue, 2);
        Commit();
    }

    void BoldChanged(object sender, RoutedEventArgs e)
    {
        if (_init) return;
        _p.BoldText = BoldSwitch.IsChecked == true;
        Commit();
    }

    void SignOut(object sender, RoutedEventArgs e)
    {
        if (!Prompt.Confirm(Window.GetWindow(this), "退出登录？", "会断开与 Halo 实例的会话并回到登录页。", "退出登录")) return;
        if (Window.GetWindow(this) is MainWindow w) w.SignOut();
    }

    void RenderAbout()
    {
        AboutVersion.Text = $"版本 {typeof(App).Assembly.GetName().Version?.ToString(3) ?? "1.0.0"}";
        AboutBackdrop.Text = WindowFx.MicaActive ? "Mica"
            : WindowFx.MicaSupported ? "纯色（Mica 未启用）"
            : "纯色（Windows 10 不支持 Mica）";
        var dpi = VisualTreeHelper.GetDpi(this);
        AboutDpi.Text = $"{dpi.DpiScaleX * 100:0}%";
        AboutOs.Text = $"Windows {Environment.OSVersion.Version.Build}";
    }
}
