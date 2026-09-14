using System.Windows.Controls;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class NetworkView : UserControl
{
    public NetworkView() => InitializeComponent();
    public void Bind(AppState s) => Tunnels.ItemsSource = s.Tunnels;
}
