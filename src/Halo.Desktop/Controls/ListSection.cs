using System.Windows;
using System.Windows.Controls;

namespace Halo.Desktop.Controls;

/// <summary>
/// Inset grouped list (macOS System Settings / iOS insetGrouped): optional 11pt label2
/// header, grouped2 card with 14pt continuous corners, optional footer.
/// Children are usually <see cref="ListRow"/>s, inline or generated from ItemsSource.
/// </summary>
public class ListSection : ItemsControl
{
    public static readonly DependencyProperty HeaderProperty = DependencyProperty.Register(
        nameof(Header), typeof(string), typeof(ListSection), new PropertyMetadata(null));

    public static readonly DependencyProperty FooterProperty = DependencyProperty.Register(
        nameof(Footer), typeof(string), typeof(ListSection), new PropertyMetadata(null));

    public string? Header { get => (string?)GetValue(HeaderProperty); set => SetValue(HeaderProperty, value); }
    public string? Footer { get => (string?)GetValue(FooterProperty); set => SetValue(FooterProperty, value); }

    protected override void OnItemsChanged(System.Collections.Specialized.NotifyCollectionChangedEventArgs e)
    {
        base.OnItemsChanged(e);
        // Rows decide their own separator from their index; nudge them after the generator settles.
        Dispatcher.BeginInvoke(() =>
        {
            foreach (var row in FindRows(this)) row.UpdateSeparator();
        }, System.Windows.Threading.DispatcherPriority.Loaded);
    }

    static IEnumerable<ListRow> FindRows(DependencyObject root)
    {
        var n = System.Windows.Media.VisualTreeHelper.GetChildrenCount(root);
        for (var i = 0; i < n; i++)
        {
            var child = System.Windows.Media.VisualTreeHelper.GetChild(root, i);
            if (child is ListRow r) { yield return r; continue; }
            foreach (var x in FindRows(child)) yield return x;
        }
    }
}
