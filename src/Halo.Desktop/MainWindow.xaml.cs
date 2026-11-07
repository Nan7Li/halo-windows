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

    protected override void OnSourceInitialized(EventArgs e)
    {
        base.OnSourceInitialized(e);
        WindowFx.Init(this, ThemeManager.IsDark);
        ApplyBackdrop();
        HwndSource.FromHwnd(new WindowInteropHelper(this).Handle)?.AddHook(WndProc);
    }

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
        _snapshot.Clear();
        ShowPage(HaloPage.Home);
        await RefreshAll();
        WarmPages();
        _poll.Start();
    }

    void LogoutClick(object sender, RoutedEventArgs e) => SignOut();

    internal void SignOut()
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
        // Data is kept current by the 4 s poll, so switching pages does not fetch anything.
        if (!IsLoaded || _showing) return;
        var page = _nav.FirstOrDefault(kv => ReferenceEquals(kv.Value, sender)).Key;
        ShowPage(page);
    }

    bool _showing;

    internal void ShowPage(HaloPage page, bool animate = true)
    {
        _s.Page = page;
        PageTitle.Text = _s.Title;
        _showing = true;
        try { if (_nav.TryGetValue(page, out var rb) && rb.IsChecked != true) rb.IsChecked = true; }
        finally { _showing = false; }
        if (_stale.Remove(page)) BindPage(page);
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

    // Which data each page shows; a page is re-bound only when one of these changed.
    static readonly Dictionary<HaloPage, string[]> PageData = new()
    {
        [HaloPage.Home] = ["devices", "host", "tasks", "threads", "calls"],
        [HaloPage.Messages] = ["threads"],
        [HaloPage.Phone] = ["devices", "calls"],
        [HaloPage.Devices] = ["devices"],
        [HaloPage.Network] = ["tunnels"],
        [HaloPage.Logs] = ["logs"],
        [HaloPage.Settings] = ["identity"]
    };

    readonly Dictionary<string, string> _snapshot = [];
    readonly HashSet<HaloPage> _stale = [];

    /// <summary>True when <paramref name="value"/> differs from the last value seen for this key.</summary>
    bool Changed(string key, object? value)
    {
        string sig;
        try { sig = System.Text.Json.JsonSerializer.Serialize(value); }
        catch { sig = Guid.NewGuid().ToString(); }
        if (_snapshot.TryGetValue(key, out var old) && old == sig) return false;
        _snapshot[key] = sig;
        return true;
    }

    /// <summary>
    /// Polls the host. Collections are only replaced when their content changed, and only the
    /// visible page is re-bound; other pages are marked stale and re-bound when shown. Re-binding
    /// regenerates every row, so doing it for all seven pages every 4 s made page switches
    /// re-layout from scratch.
    /// </summary>
    internal async Task RefreshAll()
    {
        try
        {
            var changed = new HashSet<string>();

            var devices = await _s.Client.DevicesAsync();
            if (Changed("devices", devices)) { changed.Add("devices"); _s.Replace(_s.Devices, devices); }
            if (_s.SelectedDevice == null || _s.Devices.All(d => d.Id != _s.SelectedDevice.Id))
                _s.SelectedDevice = _s.Devices.FirstOrDefault();

            var host = await _s.Client.HostAsync();
            if (Changed("host", host))
            {
                changed.Add("host");
                _s.Host.CpuModel = host.CpuModel;
                _s.Host.Board = host.Board;
                _s.Host.Cpu = host.Cpu;
                _s.Host.Memory = host.Memory;
                _s.Host.Disk = host.Disk;
                _s.Host.Rx = host.Rx;
                _s.Host.Tx = host.Tx;
                _s.Host.MemoryText = host.MemoryText;
            }

            async Task Load<T>(string key, System.Collections.ObjectModel.ObservableCollection<T> target, Task<List<T>> fetch)
            {
                var items = await fetch;
                if (Changed(key, items)) { changed.Add(key); _s.Replace(target, items); }
            }
            await Load("tasks", _s.Tasks, _s.Client.TasksAsync());
            await Load("threads", _s.Threads, _s.Client.ContactsAsync());
            await Load("calls", _s.Calls, _s.Client.CallsAsync(_s.SelectedDevice?.Id));
            await Load("logs", _s.Logs, _s.Client.LogsAsync());
            await Load("tunnels", _s.Tunnels, _s.Client.TunnelsAsync());
            if (Changed("identity", new { _s.Server, _s.Version, _s.Username, _s.Role, _s.Client.Preview })) changed.Add("identity");

            foreach (var (page, keys) in PageData)
                if (keys.Any(changed.Contains)) _stale.Add(page);
            if (_stale.Remove(_s.Page)) BindPage(_s.Page);
            PageSub.Text = _s.StatusLine;
        }
        catch (Exception ex)
        {
            PageSub.Text = "同步失败 · " + ex.Message;
        }
    }

    void BindPage(HaloPage page)
    {
        switch (page)
        {
            case HaloPage.Home: HomePage.Bind(_s); break;
            case HaloPage.Messages: MsgPage.Bind(_s); break;
            case HaloPage.Phone: PhonePage.Bind(_s); break;
            case HaloPage.Devices: DevPage.Bind(_s); break;
            case HaloPage.Network: NetPage.Bind(_s); break;
            case HaloPage.Logs: LogsPage.Bind(_s); break;
            case HaloPage.Settings: SetPage.Bind(_s); break;
        }
    }

    /// <summary>
    /// After sign-in, binds and lays out the hidden pages one at a time while the app is idle,
    /// so the first visit to each page does not pay for building its whole visual tree.
    /// </summary>
    void WarmPages()
    {
        foreach (var (page, el) in _pages)
        {
            Dispatcher.BeginInvoke(() =>
            {
                if (!_s.LoggedIn || ReferenceEquals(el, _current) || el.Visibility != Visibility.Collapsed) return;
                if (_stale.Remove(page)) BindPage(page);
                el.Visibility = Visibility.Hidden;
                el.UpdateLayout();
                if (!ReferenceEquals(el, _current) && el.Visibility == Visibility.Hidden) el.Visibility = Visibility.Collapsed;
            }, DispatcherPriority.ApplicationIdle);
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
