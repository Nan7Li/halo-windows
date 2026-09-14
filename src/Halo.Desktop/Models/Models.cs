using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Runtime.CompilerServices;
using System.Windows.Media;

namespace Halo.Desktop.Models;

public abstract class Observable : INotifyPropertyChanged
{
    public event PropertyChangedEventHandler? PropertyChanged;
    protected bool Set<T>(ref T field, T value, [CallerMemberName] string? name = null)
    {
        if (EqualityComparer<T>.Default.Equals(field, value)) return false;
        field = value;
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
        return true;
    }
    protected void Raise([CallerMemberName] string? name = null) =>
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(name));
}

public sealed class SessionInfo
{
    public bool Authenticated { get; set; }
    public string Username { get; set; } = "";
    public string Role { get; set; } = "Administrator";
    public string Csrf { get; set; } = "";
    public string Version { get; set; } = "";
}

public sealed class HostPerf : Observable
{
    public string CpuModel { get => _cpu; set => Set(ref _cpu, value); }
    public string Board { get => _board; set => Set(ref _board, value); }
    public double Cpu { get => _cpuP; set => Set(ref _cpuP, value); }
    public double Memory { get => _mem; set => Set(ref _mem, value); }
    public double Disk { get => _disk; set => Set(ref _disk, value); }
    public long Rx { get => _rx; set => Set(ref _rx, value); }
    public long Tx { get => _tx; set => Set(ref _tx, value); }
    public string MemoryText { get => _mt; set => Set(ref _mt, value); }
    string _cpu = "—", _board = "—", _mt = "—";
    double _cpuP, _mem, _disk;
    long _rx, _tx;
}

public sealed class DeviceItem : Observable
{
    public string Id { get; set; } = "";
    public string Name { get => _name; set => Set(ref _name, value); }
    public string Operator { get => _op; set => Set(ref _op, value); }
    public string Mode { get => _mode; set => Set(ref _mode, value); }
    public string Band { get => _band; set => Set(ref _band, value); }
    public string Imei { get; set; } = "";
    public string Iccid { get; set; } = "";
    public string PublicIp { get => _ip; set => Set(ref _ip, value); }
    public string Phone { get => _phone; set => Set(ref _phone, value); }
    public int SignalDbm { get => _dbm; set { if (Set(ref _dbm, value)) Raise(nameof(SignalBars)); } }
    public bool Healthy { get => _ok; set { if (Set(ref _ok, value)) Raise(nameof(StatusText)); } }
    public bool Vowifi { get => _vw; set => Set(ref _vw, value); }
    public bool ImsReady { get => _ims; set => Set(ref _ims, value); }
    public bool Flight { get => _fl; set => Set(ref _fl, value); }
    public bool Running { get => _run; set => Set(ref _run, value); }
    public string Radio { get => _radio; set => Set(ref _radio, value); }
    public string StatusText => Healthy ? "在线" : "离线";
    public Brush StatusBrush => new SolidColorBrush(Healthy ? Color.FromRgb(48, 209, 88) : Color.FromRgb(255, 69, 58));
    public int SignalBars => SignalDbm >= -85 ? 4 : SignalDbm >= -95 ? 3 : SignalDbm >= -105 ? 2 : SignalDbm >= -115 ? 1 : 0;
    public string SignalText => SignalDbm == 0 ? "—" : $"{SignalDbm} dBm";
    public string Initial => string.IsNullOrWhiteSpace(Name) ? "?" : Name[..1].ToUpperInvariant();
    string _name = "", _op = "—", _mode = "—", _band = "—", _ip = "—", _phone = "—", _radio = "cellular";
    int _dbm;
    bool _ok, _vw, _ims, _fl, _run = true;
}

public sealed class SmsThread : Observable
{
    public string Key { get; set; } = "";
    public string Peer { get; set; } = "";
    public string Display { get => _d; set => Set(ref _d, value); }
    public string Preview { get => _p; set => Set(ref _p, value); }
    public string When { get => _w; set => Set(ref _w, value); }
    public string DeviceId { get; set; } = "";
    public string DeviceName { get; set; } = "";
    public string Imei { get; set; } = "";
    public string Imsi { get; set; } = "";
    public int Unread { get => _u; set { if (Set(ref _u, value)) Raise(nameof(HasUnread)); } }
    public bool HasUnread => Unread > 0;
    public string Avatar => string.IsNullOrWhiteSpace(Display) ? "#" : Display.TrimStart('+')[..1];
    string _d = "", _p = "", _w = "";
    int _u;
}

public sealed class SmsBubble : Observable
{
    public int Id { get; set; }
    public string Body { get; set; } = "";
    public string Time { get; set; } = "";
    public bool Mine { get; set; }
    public string Status { get; set; } = "";
    public System.Windows.HorizontalAlignment Align =>
        Mine ? System.Windows.HorizontalAlignment.Right : System.Windows.HorizontalAlignment.Left;
}

public sealed class CallRow
{
    public int Id { get; set; }
    public string Number { get; set; } = "";
    public string Direction { get; set; } = "";
    public string State { get; set; } = "";
    public string When { get; set; } = "";
    public string Duration { get; set; } = "";
    public string Transport { get; set; } = "";
    public string Glyph => Direction == "incoming" ? "↙" : "↗";
}

public sealed class LogRow
{
    public string Time { get; set; } = "";
    public string Level { get; set; } = "info";
    public string Message { get; set; } = "";
    public Brush LevelBrush => new SolidColorBrush(Level.ToLowerInvariant() switch
    {
        "error" or "fatal" => Color.FromRgb(255, 69, 58),
        "warn" or "warning" => Color.FromRgb(255, 214, 10),
        "debug" => Color.FromRgb(191, 90, 242),
        _ => Color.FromRgb(100, 210, 255)
    });
}

public sealed class TaskRow
{
    public string Name { get; set; } = "";
    public string Next { get; set; } = "";
}

public sealed class TunnelRow
{
    public string Id { get; set; } = "";
    public string Name { get; set; } = "";
    public bool Running { get; set; }
    public string Endpoint { get; set; } = "";
    public string Status => Running ? "已连接" : "已断开";
}

public enum HaloPage { Home, Messages, Phone, Devices, Network, Logs, Settings }
