using System.Windows.Controls;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class HomeView : UserControl
{
    public HomeView() => InitializeComponent();

    public void Bind(AppState s)
    {
        Hello.Text = s.Greeting;
        HelloSub.Text = s.Client.Preview
            ? "这是桌面端预览，连上 Halo 实例后这里会变成真实射频数据。"
            : $"{s.Server} · {s.Version}";
        OnlineText.Text = $"{s.Online}";
        OnlineSub.Text = $"共 {s.Devices.Count} 台";
        CpuText.Text = $"{s.Host.Cpu:0}%";
        CpuBar.Value = Math.Clamp(s.Host.Cpu, 0, 100);
        MemText.Text = $"{s.Host.Memory:0}%";
        MemSub.Text = s.Host.MemoryText;
        NetText.Text = $"↓ {Kb(s.Host.Rx)}   ↑ {Kb(s.Host.Tx)}";
        BoardText.Text = string.IsNullOrWhiteSpace(s.Host.CpuModel) ? "—" : s.Host.CpuModel;
        DeviceList.ItemsSource = s.Devices;
        TaskList.ItemsSource = s.Tasks;
    }

    static string Kb(long bps) => bps < 1000 ? $"{bps} B/s" : bps < 1_000_000 ? $"{bps / 1000.0:0.#} KB/s" : $"{bps / 1_000_000.0:0.##} MB/s";
}
