using System.Windows.Controls;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class LogsView : UserControl
{
    AppState? _s;
    public LogsView() => InitializeComponent();

    public void Bind(AppState s)
    {
        _s = s;
        Apply();
    }

    void Filter(object sender, TextChangedEventArgs e) => Apply();

    void Apply()
    {
        if (_s == null) return;
        var q = FilterBox.Text.Trim();
        LogList.ItemsSource = string.IsNullOrEmpty(q)
            ? _s.Logs.ToList()
            : _s.Logs.Where(l => l.Message.Contains(q, StringComparison.OrdinalIgnoreCase) || l.Level.Contains(q, StringComparison.OrdinalIgnoreCase)).ToList();
        if (LogList.Items.Count > 0) LogList.ScrollIntoView(LogList.Items[^1]);
    }
}
