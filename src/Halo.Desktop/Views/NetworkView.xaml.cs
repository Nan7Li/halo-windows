using System.Windows;
using System.Windows.Controls;
using System.Windows.Controls.Primitives;
using Halo.Desktop.Controls;
using Halo.Desktop.Models;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class NetworkView : UserControl
{
    public NetworkView() => InitializeComponent();

    public void Bind(AppState s)
    {
        Tunnels.ItemsSource = s.Tunnels.ToList();
        var empty = s.Tunnels.Count == 0;
        Empty.Visibility = empty ? Visibility.Visible : Visibility.Collapsed;
        Scroll.Visibility = empty ? Visibility.Collapsed : Visibility.Visible;
    }

    // Halo's HTTP API (Services/HaloClient) only exposes GET /wireguard, so these actions
    // explain the limitation instead of pretending to work.
    void Unsupported(string what) =>
        ShowHint($"{what}需要在 Halo 主机上操作：当前的 Halo API 没有提供对应接口。");

    void ShowHint(string text)
    {
        HintText.Text = text;
        HintBar.Visibility = Visibility.Visible;
    }

    void ToggleTunnel(object sender, RoutedEventArgs e)
    {
        if (sender is IosSwitch sw && sw.DataContext is TunnelRow t)
        {
            sw.IsChecked = t.Running; // springs back to the real state
            Unsupported(t.Running ? $"断开“{t.Name}”" : $"连接“{t.Name}”");
        }
    }

    void MoreClick(object sender, RoutedEventArgs e)
    {
        if (sender is Button { ContextMenu: { } menu } b)
        {
            menu.PlacementTarget = b;
            menu.Placement = PlacementMode.Bottom;
            menu.DataContext = b.DataContext;
            menu.IsOpen = true;
        }
    }

    static TunnelRow? TunnelOf(object sender) =>
        (sender as MenuItem)?.Parent is ContextMenu m ? (m.DataContext ?? (m.PlacementTarget as FrameworkElement)?.DataContext) as TunnelRow : null;

    void EditTunnel(object sender, RoutedEventArgs e) => Unsupported($"修改“{TunnelOf(sender)?.Name}”");

    void DeleteTunnel(object sender, RoutedEventArgs e) => Unsupported($"删除“{TunnelOf(sender)?.Name}”");
}
