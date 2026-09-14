using System.Windows;
using System.Windows.Controls;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class DevicesView : UserControl
{
    AppState? _s;

    public DevicesView() => InitializeComponent();

    public void Bind(AppState s)
    {
        _s = s;
        List.ItemsSource = s.Devices;
        if (s.SelectedDevice != null)
            List.SelectedItem = s.Devices.FirstOrDefault(d => d.Id == s.SelectedDevice.Id);
        Render(s.SelectedDevice ?? s.Devices.FirstOrDefault());
    }

    void Pick(object sender, SelectionChangedEventArgs e)
    {
        if (List.SelectedItem is DeviceItem d && _s != null)
        {
            _s.SelectedDevice = d;
            Render(d);
        }
    }

    void Render(DeviceItem? d)
    {
        if (d == null)
        {
            DName.Text = "还没有模组";
            DOp.Text = "在 Halo 实例里添加设备后，这里会出现射频与 IMS 状态。";
            return;
        }
        DName.Text = d.Name;
        DOp.Text = $"{d.StatusText} · {d.Operator} · {d.Radio}";
        DSig.Text = d.SignalText;
        DMode.Text = string.IsNullOrWhiteSpace(d.Mode) ? "—" : d.Mode;
        DIp.Text = d.PublicIp;
        DImei.Text = "IMEI  " + (string.IsNullOrEmpty(d.Imei) ? "—" : d.Imei);
        DIccid.Text = "ICCID  " + (string.IsNullOrEmpty(d.Iccid) ? "—" : d.Iccid);
        DPhone.Text = "号码  " + d.Phone;
        DHint.Text = d.ImsReady ? "IMS 已就绪" : d.Vowifi ? "WiFi Calling 开启中" : d.Flight ? "飞行模式" : "";
    }

    async void Flight(object s, RoutedEventArgs e)
    {
        if (_s?.SelectedDevice is null) return;
        try { await _s.Client.FlightAsync(_s.SelectedDevice.Id, !_s.SelectedDevice.Flight); DHint.Text = "已发送飞行模式指令"; }
        catch (Exception ex) { DHint.Text = ex.Message; }
    }

    async void Vowifi(object s, RoutedEventArgs e)
    {
        if (_s?.SelectedDevice is null) return;
        try { await _s.Client.VowifiAsync(_s.SelectedDevice.Id, !_s.SelectedDevice.Vowifi); DHint.Text = "已发送 WiFi Calling 指令"; }
        catch (Exception ex) { DHint.Text = ex.Message; }
    }

    async void Reboot(object s, RoutedEventArgs e)
    {
        if (_s?.SelectedDevice is null) return;
        if (MessageBox.Show("重启这块模组？驻网会中断几秒到几十秒。", "Halo", MessageBoxButton.OKCancel) != MessageBoxResult.OK) return;
        try { await _s.Client.RebootAsync(_s.SelectedDevice.Id); DHint.Text = "正在重启"; }
        catch (Exception ex) { DHint.Text = ex.Message; }
    }
}
