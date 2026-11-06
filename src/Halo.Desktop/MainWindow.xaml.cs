using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using System.Windows.Interop;
using System.Windows.Media;
using System.Windows.Threading;
using Halo.Desktop.Controls;
using Halo.Desktop.Models;
using Halo.Desktop.Native;
using Halo.Desktop.Services;

namespace Halo.Desktop;

public partial class MainWindow : Window
{
    readonly AppState _s = AppState.Current;
    readonly DispatcherTimer _poll = new() { Interval = TimeSpan.FromSeconds(4) };
    readonly TranslateTransform _navTT = new();
    Dictionary<HaloPage, FrameworkElement> _pages = [];
    Dictionary<HaloPage, RadioButton> _nav = [];
    FrameworkElement? _current;
    bool _navPlaced;

    public MainWindow()
    {
        InitializeComponent();
        WindowFx.Attach(this, () => ThemeManager.IsDark);
        SourceInitialized += (_, _) =>
        {
            ApplyBackdrop();
            HwndSource.FromHwnd(new WindowInteropHelper(this).Handle)?.AddHook(WndProc);
        };
        ThemeManager.Changed += OnThemeChanged;
        Loaded += OnLoaded;
        StateChanged += (_, _) => OnStateChanged();
        _poll.Tick += async (_, _) => await RefreshQuiet();
        ServerBox.Text = _s.Prefs.Server;
        UserBox.Text = _s.Prefs.Username;
        NavIndicator.RenderTransform = _navTT;
    }

    void OnLoaded(object sender, RoutedEventArgs e)
    {
        _pages = new()
        {
            [HaloPage.Home] = HomePage,
            [HaloPage.Messages] = MsgPage,
            [HaloPage.Phone] = PhonePage,
            [HaloPage.Devices] = DevPage,
            [HaloPage.Network] = NetPage,
            [HaloPage.Logs] = LogsPage,
            [HaloPage.Settings] = SetPage
        };
        _nav = new()
        {
            [HaloPage.Home] = NavHome,
            [HaloPage.Messages] = NavMsg,
            [HaloPage.Phone] = NavPhone,
            [HaloPage.Devices] = NavDev,
            [HaloPage.Network] = NavNet,
            [HaloPage.Logs] = NavLogs,
            [HaloPage.Settings] = NavSet
        };
        _current = HomePage;
        if (string.IsNullOrEmpty(PassBox.Password)) PassBox.Focus();
    }

    // ------------------------------------------------------------ window chrome

    void ApplyBackdrop()
    {
        // Mica shows through a transparent window; Windows 10 keeps the opaque grouped colour.
        if (WindowFx.MicaActive) Background = Brushes.Transparent;
        else SetResourceReference(BackgroundProperty, "GroupedBrush");
    }

    void OnThemeChanged()
    {
        WindowFx.SetDark(this, ThemeManager.IsDark);
        ApplyBackdrop();
    }

    IntPtr WndProc(IntPtr hwnd, int msg, IntPtr wParam, IntPtr lParam, ref bool handled)
    {
        const int WM_SETTINGCHANGE = 0x001A;
        if (msg == WM_SETTINGCHANGE && lParam != IntPtr.Zero &&
            System.Runtime.InteropServices.Marshal.PtrToStringUni(lParam) == "ImmersiveColorSet" &&
            _s.Prefs.Appearance == "system")
        {
            ThemeManager.Apply(_s.Prefs);
        }
        return IntPtr.Zero;
    }

    void OnStateChanged()
    {
        // A WindowChrome window overhangs the screen by its frame when maximised.
        Root.Margin = WindowState == WindowState.Maximized ? new Thickness(7) : new Thickness(0);
        var glyph = WindowState == WindowState.Maximized ? Glyphs.Restore : Glyphs.Maximize;
        var tip = WindowState == WindowState.Maximized ? "还原" : "最大化";
        LoginMaxBtn.Content = ShellMaxBtn.Content = glyph;
        LoginMaxBtn.ToolTip = ShellMaxBtn.ToolTip = tip;
    }

    void DragArea(object sender, MouseButtonEventArgs e)
    {
        if (e.ChangedButton != MouseButton.Left) return;
        if (e.ClickCount == 2) { ToggleMax(); return; }
        try { DragMove(); } catch (InvalidOperationException) { /* mouse already released */ }
    }

    void ToggleMax() => WindowState = WindowState == WindowState.Maximized ? WindowState.Normal : WindowState.Maximized;
    void MinClick(object sender, RoutedEventArgs e) => WindowState = WindowState.Minimized;
    void MaxClick(object sender, RoutedEventArgs e) => ToggleMax();
    void CloseClick(object sender, RoutedEventArgs e) => Close();

    // ------------------------------------------------------------ login

    void LoginKey(object sender, KeyEventArgs e)
    {
        if (e.Key == Key.Enter) { e.Handled = true; LoginClick(sender, e); }
    }

    async void LoginClick(object sender, RoutedEventArgs e)
    {
        if (!LoginBtn.IsEnabled) return;
        LoginError.Visibility = Visibility.Collapsed;
        LoginBtn.IsEnabled = false;
        LoginBtn.Content = "正在连接…";
        try
        {
            await _s.Client.LoginAsync(ServerBox.Text, UserBox.Text, PassBox.Password);
            _s.Prefs.Server = ServerBox.Text.Trim();
            _s.Prefs.Username = UserBox.Text.Trim();
            _s.Prefs.Save();
            await EnterShell();
        }
        catch (Exception ex)
        {
            LoginErrorText.Text = ex.Message;
            LoginError.Visibility = Visibility.Visible;
        }
        finally
        {
            LoginBtn.IsEnabled = true;
            LoginBtn.Content = "登录";
        }
    }

    async void PreviewClick(object sender, RoutedEventArgs e)
    {
        _s.Client.EnterPreview();
        await EnterShell();
    }

    /// <summary>Used by the screenshot harness.</summary>
    internal Task EnterPreviewAsync()
    {
        _s.Client.EnterPreview();
        return EnterShell();
    }

    async Task EnterShell()
    {
        _s.LoggedIn = true;
        _s.SyncIdentity();
        UserName.Text = string.IsNullOrEmpty(_s.Username) ? "Halo" : _s.Username;
        UserRole.Text = _s.Client.Preview ? "本地预览" : _s.Role;
        UserInitial.Text = string.IsNullOrEmpty(_s.Username) ? "H" : _s.Username[..1].ToUpperInvariant();
        PageSub.Text = _s.StatusLine;
        Motion.CrossFade(LoginLayer, ShellLayer);
        ShowPage(HaloPage.Home);
        await RefreshAll();
        _poll.Start();
    }

    void LogoutClick(object sender, RoutedEventArgs e)
    {
        _poll.Stop();
        _s.Client.SignOut();
        _s.LoggedIn = false;
        PassBox.Clear();
        Motion.CrossFade(ShellLayer, LoginLayer);
    }

    void AccountMoreClick(object sender, RoutedEventArgs e)
    {
        if (AccountMore.ContextMenu is { } menu)
        {
            menu.PlacementTarget = AccountMore;
            menu.Placement = System.Windows.Controls.Primitives.PlacementMode.Top;
            menu.IsOpen = true;
        }
    }

    void OpenSettings(object sender, RoutedEventArgs e) => NavSet.IsChecked = true;

    // ------------------------------------------------------------ navigation

    void NavChecked(object sender, RoutedEventArgs e)
    {
        if (!IsLoaded) return;
        var page = _nav.FirstOrDefault(kv => ReferenceEquals(kv.Value, sender)).Key;
        ShowPage(page);
        _ = RefreshAll();
    }

    internal void ShowPage(HaloPage page, bool animate = true)
    {
        _s.Page = page;
        PageTitle.Text = _s.Title;
        if (_nav.TryGetValue(page, out var rb) && rb.IsChecked != true) rb.IsChecked = true;
        var target = _pages[page];
        if (animate) Motion.CrossFade(_current, target);
        else
        {
            foreach (var p in _pages.Values) { p.BeginAnimation(OpacityProperty, null); p.Opacity = 1; p.Visibility = Visibility.Collapsed; }
            target.Visibility = Visibility.Visible;
            target.IsHitTestVisible = true;
        }
        _current = target;
        BuildToolbar(page);
        Dispatcher.BeginInvoke(() => PlaceNavIndicator(animate && _navPlaced), DispatcherPriority.Loaded);
    }

    void PlaceNavIndicator(bool animate)
    {
        if (!_nav.TryGetValue(_s.Page, out var rb) || !rb.IsLoaded) return;
        var y = rb.TranslatePoint(new Point(0, 0), NavStack).Y;
        if (!animate)
        {
            _navTT.BeginAnimation(TranslateTransform.YProperty, null);
            _navTT.Y = y;
        }
        else
        {
            Motion.Slide(_navTT, TranslateTransform.YProperty, y);
            if (Motion.Level == MotionLevel.Reduced)
            {
                NavIndicator.Opacity = 0;
                Motion.Fade(NavIndicator, 1, 150);
            }
        }
        _navPlaced = true;
    }

    void BuildToolbar(HaloPage page)
    {
        ToolbarActions.Children.Clear();
        void Add(string glyph, string tip, RoutedEventHandler click, bool enabled = true)
        {
            var b = new Button { Style = (Style)FindResource("Btn.Glass"), Content = glyph, ToolTip = tip, IsEnabled = enabled };
            ToolTipService.SetShowOnDisabled(b, true);
            b.Click += click;
            ToolbarActions.Children.Add(b);
        }

        switch (page)
        {
            case HaloPage.Messages:
                Add(Glyphs.Compose, "新信息", (_, _) => MsgPage.NewMessage());
                break;
            case HaloPage.Network:
                Add(Glyphs.Add, "添加隧道（需在 Halo 主机上配置）", (_, _) => { }, enabled: false);
                break;
        }
        if (page != HaloPage.Settings) Add(Glyphs.Refresh, "刷新", async (_, _) => await RefreshAll());
        ToolbarGroup.Visibility = ToolbarActions.Children.Count > 0 ? Visibility.Visible : Visibility.Collapsed;
    }

    // ------------------------------------------------------------ data

    internal async Task RefreshAll()
    {
        try
        {
            var devices = await _s.Client.DevicesAsync();
            _s.Replace(_s.Devices, devices);
            if (_s.SelectedDevice == null || devices.All(d => d.Id != _s.SelectedDevice.Id))
                _s.SelectedDevice = devices.FirstOrDefault();

            var host = await _s.Client.HostAsync();
            _s.Host.CpuModel = host.CpuModel;
            _s.Host.Board = host.Board;
            _s.Host.Cpu = host.Cpu;
            _s.Host.Memory = host.Memory;
            _s.Host.Disk = host.Disk;
            _s.Host.Rx = host.Rx;
            _s.Host.Tx = host.Tx;
            _s.Host.MemoryText = host.MemoryText;

            _s.Replace(_s.Tasks, await _s.Client.TasksAsync());
            _s.Replace(_s.Threads, await _s.Client.ContactsAsync());
            _s.Replace(_s.Calls, await _s.Client.CallsAsync(_s.SelectedDevice?.Id));
            _s.Replace(_s.Logs, await _s.Client.LogsAsync());
            _s.Replace(_s.Tunnels, await _s.Client.TunnelsAsync());

            HomePage.Bind(_s);
            MsgPage.Bind(_s);
            PhonePage.Bind(_s);
            DevPage.Bind(_s);
            NetPage.Bind(_s);
            LogsPage.Bind(_s);
            SetPage.Bind(_s);
            PageSub.Text = _s.StatusLine;
        }
        catch (Exception ex)
        {
            PageSub.Text = "同步失败 · " + ex.Message;
        }
    }

    async Task RefreshQuiet()
    {
        if (!_s.LoggedIn) return;
        try { await RefreshAll(); } catch { /* keep last snapshot */ }
    }

    void SearchChanged(object sender, TextChangedEventArgs e)
    {
        var q = SearchBox.Text.Trim();
        if (!IsLoaded || !_s.LoggedIn) return;
        if (q.Length > 0 && _s.Page != HaloPage.Messages) NavMsg.IsChecked = true;
        MsgPage.Filter(q);
    }
}
