using System.Windows;
using System.Windows.Controls;
using Halo.Desktop.Controls;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class LogsView : UserControl
{
    AppState? _s;
    int _lastCount = -1;

    public LogsView()
    {
        InitializeComponent();
        LevelPick.SelectedIndex = 0;
    }

    public void Bind(AppState s)
    {
        _s = s;
        Apply(scrollToEnd: _lastCount != s.Logs.Count);
    }

    void Filter(object sender, TextChangedEventArgs e) => Apply(true);
    void LevelChanged(object sender, SelectionChangedEventArgs e) => Apply(true);

    void Apply(bool scrollToEnd)
    {
        if (_s == null) return;
        var q = FilterBox.Text.Trim();
        var kind = LevelPick.SelectedIndex switch { 1 => "info", 2 => "warn", 3 => "error", _ => null };
        var rows = _s.Logs.Where(l =>
            (kind == null || LogKindConverter.Kind(l.Level) == kind) &&
            (q.Length == 0 || l.Message.Contains(q, StringComparison.OrdinalIgnoreCase) || l.Level.Contains(q, StringComparison.OrdinalIgnoreCase)))
            .ToList();
        LogList.ItemsSource = rows;
        _lastCount = _s.Logs.Count;

        Empty.Visibility = rows.Count == 0 ? Visibility.Visible : Visibility.Collapsed;
        Empty.Title = _s.Logs.Count == 0 ? "没有日志" : "没有匹配的日志";
        Empty.Description = _s.Logs.Count == 0 ? "Halo 主机还没有返回事件。" : "换个关键词或级别试试。";
        if (scrollToEnd && rows.Count > 0) LogList.ScrollIntoView(rows[^1]);
    }

    static string Line(LogRow l) => $"[{l.Time}] {l.Level.ToUpperInvariant()} {l.Message}";

    void CopyLine(object sender, RoutedEventArgs e)
    {
        if (LogList.SelectedItem is LogRow l) TrySetClipboard(Line(l));
    }

    void CopyAll(object sender, RoutedEventArgs e)
    {
        if (LogList.ItemsSource is IEnumerable<LogRow> rows) TrySetClipboard(string.Join(Environment.NewLine, rows.Select(Line)));
    }

    static void TrySetClipboard(string text)
    {
        try { Clipboard.SetText(text); } catch { /* clipboard busy */ }
    }
}
