using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class MessagesView : UserControl
{
    AppState? _s;
    string _filter = "";

    public MessagesView() => InitializeComponent();

    public void Bind(AppState s)
    {
        _s = s;
        ApplyFilter();
        if (_s.SelectedThread != null)
            ThreadList.SelectedItem = s.Threads.FirstOrDefault(t => t.Key == _s.SelectedThread.Key);
    }

    public void Filter(string q)
    {
        _filter = q;
        ApplyFilter();
    }

    void ApplyFilter()
    {
        if (_s == null) return;
        var q = _filter.Trim();
        ThreadList.ItemsSource = string.IsNullOrEmpty(q)
            ? _s.Threads
            : _s.Threads.Where(t =>
                t.Display.Contains(q, StringComparison.OrdinalIgnoreCase) ||
                t.Peer.Contains(q, StringComparison.OrdinalIgnoreCase) ||
                t.Preview.Contains(q, StringComparison.OrdinalIgnoreCase)).ToList();
    }

    async void ThreadPicked(object sender, SelectionChangedEventArgs e)
    {
        if (_s == null || ThreadList.SelectedItem is not SmsThread thread) return;
        _s.SelectedThread = thread;
        PeerTitle.Text = thread.Display;
        PeerSub.Text = thread.Peer + (string.IsNullOrEmpty(thread.DeviceName) ? "" : " · " + thread.DeviceName);
        try
        {
            var bubbles = await _s.Client.ThreadAsync(thread);
            _s.Replace(_s.Bubbles, bubbles);
            BubbleList.ItemsSource = null;
            BubbleList.ItemsSource = _s.Bubbles;
            ChatScroll.ScrollToEnd();
            thread.Unread = 0;
        }
        catch (Exception ex)
        {
            PeerSub.Text = ex.Message;
        }
    }

    async void SendClick(object sender, RoutedEventArgs e) => await Send();

    async void ComposerKey(object sender, KeyEventArgs e)
    {
        if (e.Key == Key.Enter && Keyboard.Modifiers == ModifierKeys.None)
        {
            e.Handled = true;
            await Send();
        }
    }

    async Task Send()
    {
        if (_s?.SelectedThread is null) return;
        var text = Composer.Text.Trim();
        if (text.Length == 0) return;
        var device = _s.SelectedThread.DeviceId;
        if (string.IsNullOrEmpty(device)) device = _s.SelectedDevice?.Id ?? _s.Devices.FirstOrDefault()?.Id ?? "";
        Composer.Text = "";
        _s.Bubbles.Add(new SmsBubble { Body = text, Mine = true, Time = DateTime.Now.ToString("HH:mm"), Status = "发送中" });
        BubbleList.ItemsSource = null;
        BubbleList.ItemsSource = _s.Bubbles;
        ChatScroll.ScrollToEnd();
        try
        {
            await _s.Client.SendSmsAsync(device, _s.SelectedThread.Peer, text);
        }
        catch (Exception ex)
        {
            MessageBox.Show(ex.Message, "发送失败");
        }
    }

    void NewSms(object sender, RoutedEventArgs e)
    {
        var owner = Window.GetWindow(this);
        var phone = Prompt.Ask(owner, "新短信", "发送到哪个号码？", "+");
        if (string.IsNullOrWhiteSpace(phone) || phone == "+") return;
        var body = Prompt.Ask(owner, "新短信", "短信内容");
        if (string.IsNullOrWhiteSpace(body)) return;
        var device = _s?.SelectedDevice?.Id ?? _s?.Devices.FirstOrDefault()?.Id ?? "";
        _ = SendNew(device, phone, body);
    }

    async Task SendNew(string device, string phone, string body)
    {
        try
        {
            if (_s != null) await _s.Client.SendSmsAsync(device, phone, body);
        }
        catch (Exception ex)
        {
            MessageBox.Show(ex.Message, "发送失败");
        }
    }
}
