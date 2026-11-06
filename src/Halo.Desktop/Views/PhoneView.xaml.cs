using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using Halo.Desktop.Controls;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class PhoneView : UserControl
{
    AppState? _s;
    string _number = "";
    bool _binding;

    public PhoneView()
    {
        InitializeComponent();
        RenderNumber();
    }

    public void Bind(AppState s)
    {
        _s = s;
        _binding = true;
        var keep = DevicePick.SelectedIndex;
        DevicePick.ItemsSource = s.Devices.Select(d => $"{d.Name} · {d.Operator}").ToList();
        if (keep >= 0 && keep < s.Devices.Count) DevicePick.SelectedIndex = keep;
        else if (s.SelectedDevice != null) DevicePick.SelectedIndex = s.Devices.ToList().FindIndex(d => d.Id == s.SelectedDevice.Id);
        else if (s.Devices.Count > 0) DevicePick.SelectedIndex = 0;
        _binding = false;
        RenderIms();

        CallList.ItemsSource = s.Calls.ToList();
        var empty = s.Calls.Count == 0;
        NoCalls.Visibility = empty ? Visibility.Visible : Visibility.Collapsed;
        RecentsScroll.Visibility = empty ? Visibility.Collapsed : Visibility.Visible;
    }

    DeviceItem? Current()
    {
        if (_s == null || DevicePick.SelectedIndex < 0 || DevicePick.SelectedIndex >= _s.Devices.Count) return _s?.SelectedDevice;
        return _s.Devices[DevicePick.SelectedIndex];
    }

    void DeviceChanged(object sender, SelectionChangedEventArgs e)
    {
        if (!_binding) RenderIms();
    }

    void RenderIms()
    {
        var dev = Current();
        ImsHint.Text = dev == null ? "没有可用设备"
            : dev.ImsReady ? "IMS 已注册 · 可通过 WiFi Calling 呼叫"
            : dev.Vowifi ? "WiFi Calling 已开启，等待 IMS 注册"
            : "WiFi Calling 未开启，将尝试蜂窝网络";
        ImsDot.SetResourceReference(System.Windows.Shapes.Shape.FillProperty,
            dev == null ? "Label3Brush" : dev.ImsReady ? "GreenBrush" : dev.Vowifi ? "OrangeBrush" : "Label3Brush");
    }

    void RenderNumber()
    {
        var empty = string.IsNullOrEmpty(_number);
        NumberFace.Text = empty ? "输入号码" : _number;
        NumberFace.SetResourceReference(TextBlock.ForegroundProperty, empty ? "Label3Brush" : "LabelBrush");
        BackBtn.Visibility = empty ? Visibility.Hidden : Visibility.Visible;
    }

    void Pad(object sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string d }) _number += d;
        RenderNumber();
    }

    void PadPlus(object sender, MouseButtonEventArgs e)
    {
        _number += "+";
        RenderNumber();
        e.Handled = true;
    }

    void Backspace(object sender, RoutedEventArgs e)
    {
        if (_number.Length > 0) _number = _number[..^1];
        RenderNumber();
    }

    protected override void OnPreviewTextInput(TextCompositionEventArgs e)
    {
        base.OnPreviewTextInput(e);
        if (e.OriginalSource is TextBox) return;
        foreach (var ch in e.Text)
            if (char.IsDigit(ch) || ch is '*' or '#' or '+') _number += ch;
        RenderNumber();
    }

    async void CallClick(object sender, RoutedEventArgs e) => await Dial(_number);

    async void RowClick(object sender, RoutedEventArgs e)
    {
        if (sender is ListRow { DataContext: CallRow c })
        {
            _number = c.Number;
            RenderNumber();
            await Dial(c.Number);
        }
    }

    async Task Dial(string number)
    {
        var dev = Current();
        if (dev == null || string.IsNullOrWhiteSpace(number) || _s == null) return;
        CallBanner.Visibility = Visibility.Visible;
        CallBanner.Style = (Style)FindResource("Banner.Info");
        CallBannerText.Text = $"正在通过 {dev.Name} 呼叫 {number}…";
        try
        {
            await _s.Client.DialAsync(dev.Id, number);
            if (_s.Client.Preview) CallBannerText.Text = $"预览模式：不会真的呼叫 {number}。";
        }
        catch (Exception ex)
        {
            CallBanner.Style = (Style)FindResource("Banner.Error");
            CallBannerText.Text = "呼叫失败：" + ex.Message;
        }
    }
}
