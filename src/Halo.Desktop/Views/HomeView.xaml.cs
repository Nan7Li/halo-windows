using System.Windows;
using System.Windows.Controls;
using Halo.Desktop.Services;

namespace Halo.Desktop.Views;

public partial class HomeView : UserControl
{
    public HomeView() => InitializeComponent();

    public void Bind(AppState s)
    {
        PreviewBanner.Visibility = s.Client.Preview ? Visibility.Visible : Visibility.Collapsed;

        OnlineText.Text = $"{s.Online}";
        OnlineTotal.Text = $"/ {s.Devices.Count}";
        DeviceList.ItemsSource = s.Devices.Take(4).ToList();

        RxText.Text = "↓ " + Rate(s.Host.Rx);
        TxText.Text = "↑ " + Rate(s.Host.Tx);
        BoardText.Text = string.IsNullOrWhiteSpace(s.Host.Board) || s.Host.Board == "—" ? s.Server : s.Host.Board;

        TaskList.ItemsSource = s.Tasks.Take(4).ToList();
        NoTasks.Visibility = s.Tasks.Count == 0 ? Visibility.Visible : Visibility.Collapsed;

        CpuModel.Text = string.IsNullOrWhiteSpace(s.Host.CpuModel) ? "" : s.Host.CpuModel;
        CpuRing.Value = s.Host.Cpu;
        CpuText.Text = $"{s.Host.Cpu:0}%";
        MemRing.Value = s.Host.Memory;
        MemText.Text = $"{s.Host.Memory:0}%";
        MemSub.Text = string.IsNullOrWhiteSpace(s.Host.MemoryText) || s.Host.MemoryText == "—" ? "内存" : "内存 · " + s.Host.MemoryText;
        DiskRing.Value = s.Host.Disk;
        DiskText.Text = $"{s.Host.Disk:0}%";

        CallList.ItemsSource = s.Calls.Take(4).ToList();
        NoCalls.Visibility = s.Calls.Count == 0 ? Visibility.Visible : Visibility.Collapsed;

        ThreadList.ItemsSource = s.Threads.Take(3).ToList();
        NoThreads.Visibility = s.Threads.Count == 0 ? Visibility.Visible : Visibility.Collapsed;
    }

    static string Rate(long bps) => bps < 1000 ? $"{bps} B/s" : bps < 1_000_000 ? $"{bps / 1000.0:0.#} KB/s" : $"{bps / 1_000_000.0:0.##} MB/s";
}
