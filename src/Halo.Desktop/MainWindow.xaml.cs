using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using System.Windows.Media.Animation;
using System.Windows.Threading;
using Halo.Desktop.Models;
using Halo.Desktop.Native;
using Halo.Desktop.Services;

namespace Halo.Desktop;

public partial class MainWindow : Window
{
    readonly AppState _s = AppState.Current;
    readonly DispatcherTimer _poll = new() { Interval = TimeSpan.FromSeconds(4) };
    FrameworkElement[] _pages = [];

    public MainWindow()
    {
        InitializeComponent();
        WindowFx.Apply(this);
        Loaded += OnLoaded;
        _poll.Tick += async (_, _) => await RefreshQuiet();
        ServerBox.Text = _s.Prefs.Server;
        UserBox.Text = _s.Prefs.Username;
        SearchBox.Text = "搜索会话、模组、号码";
        SearchBox.GotFocus += (_, _) => { if (SearchBox.Text == "搜索会话、模组、号码") SearchBox.Text = ""; };
        SearchBox.LostFocus += (_, _) => { if (string.IsNullOrWhiteSpace(SearchBox.Text)) SearchBox.Text = "搜索会话、模组、号码"; };
    }

    void OnLoaded(object sender, RoutedEventArgs e)
    {
        _pages = [HomePage, MsgPage, PhonePage, DevPage, NetPage, LogsPage, SetPage];
        if (TryFindResource("LoginIn") is Storyboard sb)
        {
            Storyboard.SetTarget(sb, HaloGlow);
            sb.Begin();
        }
    }

    void DragMoveWindow(object sender, MouseButtonEventArgs e)
    {
        if (e.ChangedButton == MouseButton.Left) DragMove();
    }

    void MinClick(object sender, RoutedEventArgs e) => WindowState = WindowState.Minimized;

    void MaxClick(object sender, RoutedEventArgs e) =>
        WindowState = WindowState == WindowState.Maximized ? WindowState.Normal : WindowState.Maximized;

    void CloseClick(object sender, RoutedEventArgs e) => Close();

    async void LoginClick(object sender, RoutedEventArgs e)
    {
        LoginError.Visibility = Visibility.Collapsed;
        LoginBtn.IsEnabled = false;
        LoginBtn.Content = "连接中…";
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
            LoginError.Text = ex.Message;
            LoginError.Visibility = Visibility.Visible;
        }
        finally
        {
            LoginBtn.IsEnabled = true;
            LoginBtn.Content = "进入";
        }
    }

    async void PreviewClick(object sender, RoutedEventArgs e)
    {
        _s.Client.EnterPreview();
        await EnterShell();
    }

    async Task EnterShell()
    {
        _s.LoggedIn = true;
        _s.SyncIdentity();
        UserName.Text = _s.Username;
        UserRole.Text = _s.Role;
        UserInitial.Text = string.IsNullOrEmpty(_s.Username) ? "H" : _s.Username[..1].ToUpperInvariant();
        TopStatus.Text = _s.StatusLine;
        LoginLayer.Visibility = Visibility.Collapsed;
        ShellLayer.Visibility = Visibility.Visible;
        ShowPage(HaloPage.Home);
        await RefreshAll();
        _poll.Start();
    }

    void LogoutClick(object sender, RoutedEventArgs e)
    {
        _poll.Stop();
        _s.Client.SignOut();
        _s.LoggedIn = false;
        ShellLayer.Visibility = Visibility.Collapsed;
        LoginLayer.Visibility = Visibility.Visible;
    }

    void NavChecked(object sender, RoutedEventArgs e)
    {
        if (!IsLoaded) return;
        var page = sender == NavMsg ? HaloPage.Messages
            : sender == NavPhone ? HaloPage.Phone
            : sender == NavDev ? HaloPage.Devices
            : sender == NavNet ? HaloPage.Network
            : sender == NavLogs ? HaloPage.Logs
            : sender == NavSet ? HaloPage.Settings
            : HaloPage.Home;
        ShowPage(page);
        _ = RefreshAll();
    }

    void ShowPage(HaloPage page)
    {
        _s.Page = page;
        PageTitle.Text = _s.Title;
        PageSub.Text = _s.Subtitle;
        var map = new Dictionary<HaloPage, FrameworkElement>
        {
            [HaloPage.Home] = HomePage,
            [HaloPage.Messages] = MsgPage,
            [HaloPage.Phone] = PhonePage,
            [HaloPage.Devices] = DevPage,
            [HaloPage.Network] = NetPage,
            [HaloPage.Logs] = LogsPage,
            [HaloPage.Settings] = SetPage
        };
        foreach (var p in _pages) p.Visibility = Visibility.Collapsed;
        map[page].Visibility = Visibility.Visible;
        if (TryFindResource("PageIn") is Storyboard sb) sb.Begin();
    }

    async Task RefreshAll()
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
            TopStatus.Text = _s.StatusLine;
        }
        catch (Exception ex)
        {
            TopStatus.Text = "同步失败 · " + ex.Message;
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
        if (q is "" or "搜索会话、模组、号码") return;
        NavMsg.IsChecked = true;
        MsgPage.Filter(q);
    }
}
