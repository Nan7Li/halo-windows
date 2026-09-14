using System.Collections.ObjectModel;
using System.Windows;
using Halo.Desktop.Models;

namespace Halo.Desktop.Services;

public sealed class AppState : Observable
{
    public static AppState Current { get; } = new();

    public HaloClient Client { get; } = new();
    public Prefs Prefs { get; } = Prefs.Load();

    public bool LoggedIn { get => _in; set { if (Set(ref _in, value)) Raise(nameof(ShellVisible)); } }
    public Visibility ShellVisible => LoggedIn ? Visibility.Visible : Visibility.Collapsed;
    public Visibility LoginVisible => LoggedIn ? Visibility.Collapsed : Visibility.Visible;

    public HaloPage Page
    {
        get => _page;
        set
        {
            if (Set(ref _page, value))
            {
                Raise(nameof(Title));
                Raise(nameof(Subtitle));
            }
        }
    }

    public string Title => Page switch
    {
        HaloPage.Home => "今天",
        HaloPage.Messages => "消息",
        HaloPage.Phone => "电话",
        HaloPage.Devices => "模组",
        HaloPage.Network => "链路",
        HaloPage.Logs => "现场",
        HaloPage.Settings => "设置",
        _ => "Halo"
    };

    public string Subtitle => Page switch
    {
        HaloPage.Home => "模组、信号与出口，一眼看完",
        HaloPage.Messages => "短信会话，像一条对话那样处理",
        HaloPage.Phone => "WiFi Calling 拨号与近期通话",
        HaloPage.Devices => "射频、飞行模式、IMS",
        HaloPage.Network => "出口代理与 WG 隧道",
        HaloPage.Logs => "硬件与业务事件",
        HaloPage.Settings => "这台电脑上的 Halo 客户端",
        _ => ""
    };

    public string Username { get => _user; set => Set(ref _user, value); }
    public string Role { get => _role; set => Set(ref _role, value); }
    public string Version { get => _ver; set => Set(ref _ver, value); }
    public string Server { get => _srv; set => Set(ref _srv, value); }
    public string StatusLine { get => _st; set => Set(ref _st, value); }
    public bool Busy { get => _busy; set => Set(ref _busy, value); }

    public ObservableCollection<DeviceItem> Devices { get; } = [];
    public ObservableCollection<SmsThread> Threads { get; } = [];
    public ObservableCollection<SmsBubble> Bubbles { get; } = [];
    public ObservableCollection<CallRow> Calls { get; } = [];
    public ObservableCollection<LogRow> Logs { get; } = [];
    public ObservableCollection<TaskRow> Tasks { get; } = [];
    public ObservableCollection<TunnelRow> Tunnels { get; } = [];
    public HostPerf Host { get; } = new();

    public DeviceItem? SelectedDevice { get => _dev; set => Set(ref _dev, value); }
    public SmsThread? SelectedThread { get => _th; set => Set(ref _th, value); }

    public int Online => Devices.Count(d => d.Healthy);
    public string Greeting
    {
        get
        {
            var h = DateTime.Now.Hour;
            var hello = h < 5 ? "夜深了" : h < 12 ? "早上好" : h < 18 ? "下午好" : "晚上好";
            return string.IsNullOrEmpty(Username) ? hello : $"{hello}，{Username}";
        }
    }

    string _user = "", _role = "Administrator", _ver = "", _srv = "", _st = "未连接";
    bool _in, _busy;
    HaloPage _page = HaloPage.Home;
    DeviceItem? _dev;
    SmsThread? _th;

    public void SyncIdentity()
    {
        Username = Client.Username;
        Role = Client.Role;
        Version = Client.Version;
        Server = Client.Preview ? "本地预览" : Client.BaseUrl;
        StatusLine = Client.Preview ? "预览模式 · 未连接实例" : "已连接 " + Client.BaseUrl;
        Raise(nameof(Greeting));
        Raise(nameof(Online));
    }

    public void Replace<T>(ObservableCollection<T> target, IEnumerable<T> items)
    {
        target.Clear();
        foreach (var item in items) target.Add(item);
        Raise(nameof(Online));
    }
}
