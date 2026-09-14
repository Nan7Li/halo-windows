using System.Windows;
using System.Windows.Controls;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class PhoneView : UserControl
{
    AppState? _s;
    string _number = "";

    public PhoneView() => InitializeComponent();

    public void Bind(AppState s)
    {
        _s = s;
        DevicePick.ItemsSource = s.Devices.Select(d => d.Name + "  ·  " + d.Operator).ToList();
        if (s.SelectedDevice != null)
        {
            var i = s.Devices.ToList().FindIndex(d => d.Id == s.SelectedDevice.Id);
            if (i >= 0) DevicePick.SelectedIndex = i;
        }
        else if (DevicePick.Items.Count > 0 && DevicePick.SelectedIndex < 0)
            DevicePick.SelectedIndex = 0;
        var dev = Current();
        ImsHint.Text = dev == null ? "选择一台模组"
            : dev.ImsReady ? "IMS 已注册 · 可以走 WiFi Calling"
            : dev.Vowifi ? "WiFi Calling 已开，等待 IMS"
            : "当前未开 WiFi Calling，仍可尝试蜂窝侧";
        CallList.ItemsSource = s.Calls;
        NumberFace.Text = string.IsNullOrEmpty(_number) ? "输入号码" : _number;
        NumberFace.Opacity = string.IsNullOrEmpty(_number) ? 0.35 : 1;
    }

    Models.DeviceItem? Current()
    {
        if (_s == null || DevicePick.SelectedIndex < 0 || DevicePick.SelectedIndex >= _s.Devices.Count) return _s?.SelectedDevice;
        return _s.Devices[DevicePick.SelectedIndex];
    }

    void Pad(object sender, RoutedEventArgs e)
    {
        if (sender is Button b) _number += b.Content?.ToString() ?? "";
        NumberFace.Text = _number;
        NumberFace.Opacity = 1;
    }

    void Backspace(object sender, RoutedEventArgs e)
    {
        if (_number.Length > 0) _number = _number[..^1];
        NumberFace.Text = string.IsNullOrEmpty(_number) ? "输入号码" : _number;
        NumberFace.Opacity = string.IsNullOrEmpty(_number) ? 0.35 : 1;
    }

    async void CallClick(object sender, RoutedEventArgs e)
    {
        var dev = Current();
        if (dev == null || string.IsNullOrWhiteSpace(_number) || _s == null) return;
        try { await _s.Client.DialAsync(dev.Id, _number); }
        catch (Exception ex) { MessageBox.Show(ex.Message, "呼叫"); }
    }

    void Redial(object sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string n })
        {
            _number = n;
            NumberFace.Text = n;
            NumberFace.Opacity = 1;
        }
    }
}
