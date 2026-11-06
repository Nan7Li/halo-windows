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
    bool _binding;
    string? _openKey;

    public MessagesView()
    {
        InitializeComponent();
        ShowConversation(false);
    }

    public void Bind(AppState s)
    {
        _s = s;
        ApplyFilter();
    }

    /// <summary>Screenshot harness: open the first conversation.</summary>
    internal void SelectFirst()
    {
        if (ThreadList.Items.Count > 0 && ThreadList.SelectedIndex < 0) ThreadList.SelectedIndex = 0;
    }

    /// <summary>Called by the sidebar search field.</summary>
    public void Filter(string q)
    {
        if (FilterBox.Text != q) FilterBox.Text = q;
        _filter = q;
        ApplyFilter();
    }

    void FilterChanged(object sender, TextChangedEventArgs e)
    {
        _filter = FilterBox.Text;
        ApplyFilter();
    }

    void ApplyFilter()
    {
        if (_s == null) return;
        var q = _filter.Trim();
        var items = string.IsNullOrEmpty(q)
            ? _s.Threads.ToList()
            : _s.Threads.Where(t =>
                t.Display.Contains(q, StringComparison.OrdinalIgnoreCase) ||
                t.Peer.Contains(q, StringComparison.OrdinalIgnoreCase) ||
                t.Preview.Contains(q, StringComparison.OrdinalIgnoreCase)).ToList();

        // Rebinding on every poll must not reload the open conversation or jump its scroll.
        _binding = true;
        ThreadList.ItemsSource = items;
        if (_openKey != null)
            ThreadList.SelectedItem = items.FirstOrDefault(t => t.Key == _openKey);
        _binding = false;

        NoThreads.Visibility = items.Count == 0 ? Visibility.Visible : Visibility.Collapsed;
        NoThreads.Title = string.IsNullOrEmpty(q) ? "没有信息" : "没有结果";
        NoThreads.Description = string.IsNullOrEmpty(q) ? "收到的短信会出现在这里。" : $"没有与“{q}”匹配的会话。";
    }

    void ShowConversation(bool on)
    {
        var v = on ? Visibility.Visible : Visibility.Collapsed;
        PeerHeader.Visibility = PeerLine.Visibility = ChatScroll.Visibility = ComposerRow.Visibility = v;
        NoSelection.Visibility = on ? Visibility.Collapsed : Visibility.Visible;
    }

    async void ThreadPicked(object sender, SelectionChangedEventArgs e)
    {
        if (_binding || _s == null || ThreadList.SelectedItem is not SmsThread thread) return;
        _s.SelectedThread = thread;
        _openKey = thread.Key;
        PeerTitle.Text = thread.Display;
        PeerSub.Text = thread.Peer + (string.IsNullOrEmpty(thread.DeviceName) ? "" : " · " + thread.DeviceName);
        ShowConversation(true);
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
        // Enter sends, Shift+Enter inserts a line break.
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
        var bubble = new SmsBubble { Body = text, Mine = true, Time = DateTime.Now.ToString("HH:mm"), Status = "发送中" };
        _s.Bubbles.Add(bubble);
        Rebind();
        try
        {
            await _s.Client.SendSmsAsync(device, _s.SelectedThread.Peer, text);
            bubble.Status = "已发送";
        }
        catch (Exception ex)
        {
            bubble.Status = "未发送";
            Prompt.Alert(Window.GetWindow(this), "发送失败", ex.Message);
        }
        Rebind();
    }

    void Rebind()
    {
        BubbleList.ItemsSource = null;
        BubbleList.ItemsSource = _s?.Bubbles;
        ChatScroll.ScrollToEnd();
    }

    public void NewMessage()
    {
        var owner = Window.GetWindow(this);
        var phone = Prompt.Ask(owner, "新信息", "发送到哪个号码？", "+");
        if (string.IsNullOrWhiteSpace(phone) || phone == "+") return;
        var body = Prompt.Ask(owner, "新信息", "短信内容");
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
            Prompt.Alert(Window.GetWindow(this), "发送失败", ex.Message);
        }
    }

    static SmsThread? ThreadOf(object sender) =>
        (sender as MenuItem)?.Parent is ContextMenu { PlacementTarget: ListBoxItem { DataContext: SmsThread t } } ? t : null;

    void MarkRead(object sender, RoutedEventArgs e)
    {
        if (ThreadOf(sender) is { } t) t.Unread = 0;
    }

    void CopyNumber(object sender, RoutedEventArgs e)
    {
        if (ThreadOf(sender) is { } t)
        {
            try { Clipboard.SetText(t.Peer); } catch { /* clipboard busy */ }
        }
    }
}
