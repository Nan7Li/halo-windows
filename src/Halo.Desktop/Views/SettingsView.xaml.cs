using System.Windows.Controls;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class SettingsView : UserControl
{
    public SettingsView() => InitializeComponent();

    public void Bind(AppState s)
    {
        About.Text = $"Halo Desktop 1.0.0\n{s.Version}\nWindows 原生客户端 · .NET 8 WPF";
        Conn.Text = s.Client.Preview
            ? "当前是本地预览，没有把操作发到真实实例。"
            : $"已登录 {s.Username}（{s.Role}）\n{s.Server}";
    }
}
