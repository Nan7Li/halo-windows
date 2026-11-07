using System.Windows;
using System.Windows.Controls;
using System.Windows.Shapes;
using Halo.Desktop.Controls;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class DevicesView : UserControl
{
    AppState? _s;
    bool _binding;
    string? _shownId;

    public DevicesView() => InitializeComponent();

    public void Bind(AppState s)
    {
        _s = s;
        _binding = true;
        List.ItemsSource = s.Devices;
        if (s.SelectedDevice != null)
            List.SelectedItem = s.Devices.FirstOrDefault(d => d.Id == s.SelectedDevice.Id);
        _binding = false;
        NoDevices.Visibility = s.Devices.Count == 0 ? Visibility.Visible : Visibility.Collapsed;
        Render(s.SelectedDevice ?? s.Devices.FirstOrDefault());
    }

    void Pick(object sender, SelectionChangedEventArgs e)
    {
        if (_binding) return;
        if (List.SelectedItem is DeviceItem d && _s != null)
        {
            _s.SelectedDevice = d;
            Render(d);
        }
    }

    static string Or(string? v) => string.IsNullOrWhiteSpace(v) ? "—" : v;

    void Render(DeviceItem? d)
    {
        DetailScroll.Visibility = d == null ? Visibility.Collapsed : Visibility.Visible;
        NoDetail.Visibility = d == null ? Visibility.Visible : Visibility.Collapsed;
        if (d == null) return;

        if (_shownId != d.Id)
        {
            // New device: hide sensitive values again and clear the last command message.
            IImei.IsRevealed = IIccid.IsRevealed = false;
            HintBar.Visibility = Visibility.Collapsed;
            _shownId = d.Id;
        }

        DName.Text = d.Name;
        DStatus.Text = $"{d.StatusText} · {Or(d.Operator)} · {Or(d.Mode)}";
        DDot.SetResourceReference(Shape.FillProperty, d.Healthy ? "GreenBrush" : "RedBrush");

        TileWifi.IsChecked = d.Vowifi;
        Ui.SetDetail(TileWifi, d.ImsReady ? "IMS 已注册" : d.Vowifi ? "已开启" : "已关闭");
        TileFlight.IsChecked = d.Flight;
        Ui.SetDetail(TileFlight, d.Flight ? "已开启" : "已关闭");

        IOp.Text = Or(d.Operator);
        IPhone.Text = Or(d.Phone);
        IMode.Text = Or(d.Mode);
        IBand.Text = Or(d.Band);
        ISig.Text = d.SignalText;
        IIp.Text = Or(d.PublicIp);
        IImei.Text = Or(d.Imei);
        IIccid.Text = Or(d.Iccid);
    }

    void Hint(string text, bool error = false)
    {
        HintBar.Style = (Style)FindResource(error ? "Banner.Error" : "Banner.Info");
        HintText.Text = text;
        HintBar.Visibility = Visibility.Visible;
    }

    async void Flight(object s, RoutedEventArgs e)
    {
        if (_s?.SelectedDevice is not { } d) return;
        var on = TileFlight.IsChecked == true;
        try
        {
            await _s.Client.FlightAsync(d.Id, on);
            Hint(on ? "已发送开启飞行模式的指令。" : "已发送关闭飞行模式的指令。");
        }
        catch (Exception ex)
        {
            TileFlight.IsChecked = !on;
            Hint(ex.Message, error: true);
        }
    }

    async void Vowifi(object s, RoutedEventArgs e)
    {
        if (_s?.SelectedDevice is not { } d) return;
        var on = TileWifi.IsChecked == true;
        try
        {
            await _s.Client.VowifiAsync(d.Id, on);
            Hint(on ? "已发送开启 VoWiFi 的指令。" : "已发送关闭 VoWiFi 的指令。");
        }
        catch (Exception ex)
        {
            TileWifi.IsChecked = !on;
            Hint(ex.Message, error: true);
        }
    }

    async void Reboot(object s, RoutedEventArgs e)
    {
        if (_s?.SelectedDevice is not { } d) return;
        if (!Prompt.Confirm(Window.GetWindow(this), $"重启“{d.Name}”？", "驻网会中断几秒到几十秒，进行中的通话会挂断。", "重启", destructive: true)) return;
        try { await _s.Client.RebootAsync(d.Id); Hint("正在重启…"); }
        catch (Exception ex) { Hint(ex.Message, error: true); }
    }
}
