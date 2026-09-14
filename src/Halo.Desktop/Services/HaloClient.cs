using System.Net;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using Halo.Desktop.Models;

namespace Halo.Desktop.Services;

public sealed class HaloClient
{
    readonly CookieContainer _cookies = new();
    readonly HttpClient _http;
    static readonly JsonSerializerOptions JsonOpt = new() { PropertyNameCaseInsensitive = true };

    public string BaseUrl { get; private set; } = "";
    public string Csrf { get; private set; } = "";
    public string Username { get; private set; } = "";
    public string Role { get; private set; } = "Administrator";
    public string Version { get; private set; } = "";
    public bool Preview { get; private set; }
    public bool SignedIn { get; private set; }

    public HaloClient()
    {
        var handler = new HttpClientHandler
        {
            CookieContainer = _cookies,
            UseCookies = true,
            AutomaticDecompression = DecompressionMethods.All,
            ServerCertificateCustomValidationCallback = (_, _, _, _) => true
        };
        _http = new HttpClient(handler) { Timeout = TimeSpan.FromSeconds(12) };
        _http.DefaultRequestHeaders.TryAddWithoutValidation("Accept", "application/json");
        _http.DefaultRequestHeaders.TryAddWithoutValidation("User-Agent", "HaloDesktop/1.0");
    }

    public void EnterPreview()
    {
        Preview = true;
        SignedIn = true;
        Username = "Nan7Li";
        Role = "Administrator";
        Version = "Halo 1.0 · Preview";
        BaseUrl = "preview://local";
    }

    public void SignOut()
    {
        SignedIn = false;
        Preview = false;
        Csrf = "";
        Username = "";
        try { _ = Post("/auth/logout", new { }); } catch { /* ignore */ }
    }

    public async Task LoginAsync(string url, string user, string password, CancellationToken ct = default)
    {
        BaseUrl = url.Trim().TrimEnd('/');
        if (!BaseUrl.StartsWith("http", StringComparison.OrdinalIgnoreCase))
            BaseUrl = "http://" + BaseUrl;
        Preview = false;
        var node = await SendAsync(HttpMethod.Post, "/auth/login", new { username = user, password }, ct);
        ApplySession(node);
        SignedIn = true;
        Username = Str(node, "username", "user") is { Length: > 0 } n ? n : user;
        Role = Str(node, "role") ?? "Administrator";
        try
        {
            var info = await Get("/system/info", ct);
            Version = Str(info, "version") ?? "Halo";
        }
        catch { Version = "Halo"; }
    }

    public async Task RefreshSessionAsync(CancellationToken ct = default)
    {
        if (Preview) return;
        var node = await Get("/auth/session", ct);
        ApplySession(node);
        SignedIn = Bool(node, "authenticated") ?? true;
        Username = Str(node, "username") ?? Username;
    }

    public async Task<HostPerf> HostAsync(CancellationToken ct = default)
    {
        if (Preview) return Demo.Host();
        var n = await Get("/dashboard/host", ct);
        var host = n.TryGetProperty("host", out var h) ? h : n;
        var perf = n.TryGetProperty("perf", out var p) ? p : n;
        var used = Long(perf, "memoryUsedBytes") ?? 0;
        var total = Long(perf, "memoryTotalBytes") ?? 0;
        return new HostPerf
        {
            CpuModel = Str(host, "cpuModel") ?? "—",
            Board = Str(host, "boardModel") ?? "—",
            Cpu = Num(perf, "cpuPercent") ?? 0,
            Memory = Num(perf, "memoryPercent") ?? 0,
            Disk = Num(perf, "diskPercent") ?? 0,
            Rx = Long(perf, "netRxBps") ?? 0,
            Tx = Long(perf, "netTxBps") ?? 0,
            MemoryText = total > 0 ? $"{Bytes(used)} / {Bytes(total)}" : "—"
        };
    }

    public async Task<List<DeviceItem>> DashboardDevicesAsync(CancellationToken ct = default)
    {
        if (Preview) return Demo.Devices();
        var n = await Get("/dashboard/devices", ct);
        return ReadDevices(n);
    }

    public async Task<List<DeviceItem>> DevicesAsync(CancellationToken ct = default)
    {
        if (Preview) return Demo.Devices();
        var n = await Get("/devices", ct);
        var list = n.ValueKind == JsonValueKind.Object && n.TryGetProperty("devices", out var d) ? d : n;
        return ReadDevices(list);
    }

    public async Task<List<TaskRow>> TasksAsync(CancellationToken ct = default)
    {
        if (Preview) return Demo.Tasks();
        try
        {
            var n = await Get("/automatic-tasks", ct);
            var arr = n.ValueKind == JsonValueKind.Object && n.TryGetProperty("tasks", out var t) ? t : n;
            var rows = new List<TaskRow>();
            if (arr.ValueKind != JsonValueKind.Array) return rows;
            foreach (var item in arr.EnumerateArray())
            {
                if (!(Bool(item, "enabled") ?? false)) continue;
                rows.Add(new TaskRow
                {
                    Name = Str(item, "name") ?? "任务",
                    Next = RelTime(Str(item, "nextRunAt", "next_run_at"))
                });
            }
            return rows.Take(5).ToList();
        }
        catch { return []; }
    }

    public async Task<List<SmsThread>> ContactsAsync(string? deviceId = null, CancellationToken ct = default)
    {
        if (Preview) return Demo.Threads();
        var q = "/sms/contacts?limit=200";
        if (!string.IsNullOrWhiteSpace(deviceId) && deviceId != "all")
            q += "&device_id=" + Uri.EscapeDataString(deviceId);
        var n = await Get(q, ct);
        var arr = n.ValueKind == JsonValueKind.Array ? n : (n.TryGetProperty("contacts", out var c) ? c : n);
        var list = new List<SmsThread>();
        if (arr.ValueKind != JsonValueKind.Array) return list;
        foreach (var item in arr.EnumerateArray())
        {
            var peer = Str(item, "peer") ?? "";
            var name = Str(item, "displayName", "display_name");
            list.Add(new SmsThread
            {
                Key = $"{Str(item, "deviceId", "device_id")}|{peer}",
                Peer = peer,
                Display = string.IsNullOrWhiteSpace(name) ? peer : name,
                Preview = Str(item, "lastMessage", "lastContent", "last_message", "last_content") ?? "",
                When = RelTime(Str(item, "lastTimestamp", "last_timestamp")),
                DeviceId = Str(item, "deviceId", "device_id") ?? "",
                DeviceName = Str(item, "deviceName", "device_name") ?? "",
                Imei = Str(item, "modemImei", "modem_imei") ?? "",
                Imsi = Str(item, "imsi") ?? "",
                Unread = (int)(Num(item, "unreadCount", "unread_count") ?? 0)
            });
        }
        return list;
    }

    public async Task<List<SmsBubble>> ThreadAsync(SmsThread thread, CancellationToken ct = default)
    {
        if (Preview) return Demo.Bubbles(thread.Peer);
        var q = $"/sms/thread?peer={Uri.EscapeDataString(thread.Peer)}&limit=80";
        if (!string.IsNullOrWhiteSpace(thread.DeviceId))
            q += "&device_id=" + Uri.EscapeDataString(thread.DeviceId);
        var n = await Get(q, ct);
        var arr = n.ValueKind == JsonValueKind.Array ? n : n;
        var list = new List<SmsBubble>();
        if (arr.ValueKind != JsonValueKind.Array) return list;
        foreach (var item in arr.EnumerateArray())
        {
            var dir = (Str(item, "direction") ?? "").ToLowerInvariant();
            list.Add(new SmsBubble
            {
                Id = (int)(Num(item, "id") ?? 0),
                Body = Str(item, "body", "content") ?? "",
                Time = Clock(Str(item, "timestamp")),
                Mine = dir is "outbound" or "sent",
                Status = Str(item, "status", "deliveryState", "delivery_state") ?? ""
            });
        }
        return list;
    }

    public Task SendSmsAsync(string deviceId, string phone, string message, CancellationToken ct = default)
    {
        if (Preview) return Task.CompletedTask;
        return Post("/sms/send", new { device_id = deviceId, phone, message }, ct);
    }

    public async Task<List<CallRow>> CallsAsync(string? deviceId = null, CancellationToken ct = default)
    {
        if (Preview) return Demo.Calls();
        var q = "/calls/history?limit=80";
        if (!string.IsNullOrWhiteSpace(deviceId)) q += "&device_id=" + Uri.EscapeDataString(deviceId);
        try
        {
            var n = await Get(q, ct);
            var arr = n.TryGetProperty("records", out var r) ? r : n;
            var list = new List<CallRow>();
            if (arr.ValueKind != JsonValueKind.Array) return list;
            foreach (var item in arr.EnumerateArray())
            {
                var sec = (int)(Num(item, "durationSeconds", "duration_seconds") ?? 0);
                list.Add(new CallRow
                {
                    Id = (int)(Num(item, "id") ?? 0),
                    Number = Str(item, "number") ?? "未知号码",
                    Direction = Str(item, "direction") ?? "outgoing",
                    State = Str(item, "state") ?? "",
                    When = RelTime(Str(item, "startedAt", "started_at")),
                    Duration = sec <= 0 ? "—" : $"{sec / 60}:{sec % 60:00}",
                    Transport = Str(item, "transport") ?? "vowifi"
                });
            }
            return list;
        }
        catch { return []; }
    }

    public async Task<List<LogRow>> LogsAsync(CancellationToken ct = default)
    {
        if (Preview) return Demo.Logs();
        var n = await Get("/logs/history?lines=400", ct);
        var arr = n.ValueKind == JsonValueKind.Array ? n : (n.TryGetProperty("logs", out var l) ? l : n);
        var list = new List<LogRow>();
        if (arr.ValueKind != JsonValueKind.Array) return list;
        foreach (var item in arr.EnumerateArray())
        {
            var msg = Str(item, "message") ?? "";
            if (msg.Equals("http request", StringComparison.OrdinalIgnoreCase)) continue;
            list.Add(new LogRow
            {
                Time = Clock(Str(item, "time")),
                Level = Str(item, "level") ?? "info",
                Message = msg
            });
        }
        return list;
    }

    public async Task<List<TunnelRow>> TunnelsAsync(CancellationToken ct = default)
    {
        if (Preview) return Demo.Tunnels();
        try
        {
            var n = await Get("/wireguard", ct);
            var arr = n.TryGetProperty("tunnels", out var t) ? t : n;
            var list = new List<TunnelRow>();
            if (arr.ValueKind != JsonValueKind.Array) return list;
            foreach (var item in arr.EnumerateArray())
            {
                list.Add(new TunnelRow
                {
                    Id = Str(item, "id") ?? "",
                    Name = Str(item, "name") ?? "隧道",
                    Running = Bool(item, "running") ?? false,
                    Endpoint = Str(item, "endpoint") ?? ""
                });
            }
            return list;
        }
        catch { return []; }
    }

    public Task RebootAsync(string id, CancellationToken ct = default) =>
        Preview ? Task.CompletedTask : Post($"/devices/{Uri.EscapeDataString(id)}/reboot", new { }, ct);

    public Task FlightAsync(string id, bool on, CancellationToken ct = default) =>
        Preview ? Task.CompletedTask : Post($"/devices/{Uri.EscapeDataString(id)}/airplane", new { enabled = on }, ct);

    public Task VowifiAsync(string id, bool on, CancellationToken ct = default) =>
        Preview ? Task.CompletedTask : Post($"/devices/{Uri.EscapeDataString(id)}/vowifi", new { enabled = on }, ct);

    public Task DialAsync(string id, string number, CancellationToken ct = default) =>
        Preview ? Task.CompletedTask : Post($"/devices/{Uri.EscapeDataString(id)}/call", new { number }, ct);

    public Task HangupAsync(string id, CancellationToken ct = default) =>
        Preview ? Task.CompletedTask : Post($"/devices/{Uri.EscapeDataString(id)}/hangup", new { }, ct);

    List<DeviceItem> ReadDevices(JsonElement n)
    {
        var list = new List<DeviceItem>();
        if (n.ValueKind != JsonValueKind.Array) return list;
        foreach (var item in n.EnumerateArray())
        {
            var modem = item.TryGetProperty("modem", out var m) ? m : item;
            var vw = item.TryGetProperty("vowifiRuntime", out var vr) ? vr
                : item.TryGetProperty("vowifi_runtime", out var vr2) ? vr2 : default;
            list.Add(new DeviceItem
            {
                Id = Str(item, "id") ?? "",
                Name = Str(item, "name") ?? "模组",
                Operator = Str(modem, "operator") ?? "—",
                Mode = Str(modem, "networkMode", "network_mode") ?? "—",
                Band = Str(modem, "radioBand", "radio_band") ?? "—",
                Imei = Str(modem, "imei") ?? "",
                Iccid = Str(modem, "iccid") ?? "",
                PublicIp = Str(item, "publicIp", "public_ip") ?? "—",
                Phone = Str(modem, "phoneNumber", "phone_number") ?? "—",
                SignalDbm = (int)(Num(modem, "signalDbm", "signal_dbm") ?? 0),
                Healthy = Bool(item, "healthy") ?? false,
                Vowifi = Bool(item, "vowifiEnabled", "vowifi_enabled", "vowifiActive", "vowifi_active") ?? false,
                ImsReady = vw.ValueKind == JsonValueKind.Object && (Bool(vw, "imsReady", "ims_ready") ?? false),
                Flight = Bool(item, "flightMode", "flight_mode") ?? false,
                Running = Bool(item, "running") ?? true,
                Radio = Str(item, "radioMode", "radio_mode") ?? "cellular"
            });
        }
        return list;
    }

    void ApplySession(JsonElement node)
    {
        var token = Str(node, "csrfToken", "csrf_token");
        if (!string.IsNullOrWhiteSpace(token)) Csrf = token!;
        if (node.TryGetProperty("user", out var user) && user.ValueKind == JsonValueKind.Object)
            Username = Str(user, "username") ?? Username;
    }

    Task<JsonElement> Get(string path, CancellationToken ct = default) =>
        SendAsync(HttpMethod.Get, path, null, ct);

    Task<JsonElement> Post(string path, object body, CancellationToken ct = default) =>
        SendAsync(HttpMethod.Post, path, body, ct);

    async Task<JsonElement> SendAsync(HttpMethod method, string path, object? body, CancellationToken ct)
    {
        var uri = path.StartsWith("http", StringComparison.OrdinalIgnoreCase)
            ? path
            : BaseUrl + (path.StartsWith("/api") ? path : "/api" + path);
        using var req = new HttpRequestMessage(method, uri);
        if (body != null)
            req.Content = new StringContent(JsonSerializer.Serialize(body), Encoding.UTF8, "application/json");
        if (method != HttpMethod.Get && !string.IsNullOrEmpty(Csrf))
            req.Headers.TryAddWithoutValidation("X-CSRF-Token", Csrf);
        using var resp = await _http.SendAsync(req, ct).ConfigureAwait(false);
        var text = await resp.Content.ReadAsStringAsync(ct).ConfigureAwait(false);
        JsonElement root;
        try { root = JsonDocument.Parse(string.IsNullOrWhiteSpace(text) ? "{}" : text).RootElement.Clone(); }
        catch { root = JsonDocument.Parse("{\"message\":" + JsonSerializer.Serialize(text) + "}").RootElement.Clone(); }
        if (!resp.IsSuccessStatusCode)
        {
            var msg = Str(root, "message", "error") ?? $"HTTP {(int)resp.StatusCode}";
            JsonElement err = default;
            if (root.TryGetProperty("error", out var errNode) && errNode.ValueKind == JsonValueKind.Object)
            {
                err = errNode;
                msg = Str(err, "message", "error") ?? msg;
            }
            if (resp.StatusCode == HttpStatusCode.Forbidden && (Str(root, "code") == "invalid_csrf" || (err.ValueKind == JsonValueKind.Object && Str(err, "code") == "invalid_csrf")))
            {
                try
                {
                    var sess = await Get("/auth/session", ct);
                    ApplySession(sess);
                    if (!string.IsNullOrEmpty(Csrf)) return await SendAsync(method, path, body, ct);
                }
                catch { /* fall through */ }
            }
            throw new HaloException(msg, (int)resp.StatusCode);
        }
        return root.TryGetProperty("data", out var data) ? data.Clone() : root;
    }

    static string? Str(JsonElement n, params string[] keys)
    {
        if (n.ValueKind != JsonValueKind.Object) return null;
        foreach (var k in keys)
        {
            if (n.TryGetProperty(k, out var v) && v.ValueKind is JsonValueKind.String)
                return v.GetString();
            var snake = ToSnake(k);
            if (snake != k && n.TryGetProperty(snake, out v) && v.ValueKind is JsonValueKind.String)
                return v.GetString();
        }
        return null;
    }

    static double? Num(JsonElement n, params string[] keys)
    {
        if (n.ValueKind != JsonValueKind.Object) return null;
        foreach (var k in keys.Concat(keys.Select(ToSnake)))
        {
            if (!n.TryGetProperty(k, out var v)) continue;
            if (v.ValueKind == JsonValueKind.Number) return v.GetDouble();
            if (v.ValueKind == JsonValueKind.String && double.TryParse(v.GetString(), out var d)) return d;
        }
        return null;
    }

    static long? Long(JsonElement n, params string[] keys) =>
        Num(n, keys) is { } d ? (long)d : null;

    static bool? Bool(JsonElement n, params string[] keys)
    {
        if (n.ValueKind != JsonValueKind.Object) return null;
        foreach (var k in keys.Concat(keys.Select(ToSnake)))
        {
            if (!n.TryGetProperty(k, out var v)) continue;
            if (v.ValueKind is JsonValueKind.True or JsonValueKind.False) return v.GetBoolean();
        }
        return null;
    }

    static string ToSnake(string k)
    {
        var sb = new StringBuilder();
        foreach (var c in k)
        {
            if (char.IsUpper(c) && sb.Length > 0) sb.Append('_');
            sb.Append(char.ToLowerInvariant(c));
        }
        return sb.ToString();
    }

    static string Bytes(long n)
    {
        if (n < 1024) return $"{n} B";
        if (n < 1024 * 1024) return $"{n / 1024.0:0.#} KB";
        if (n < 1024L * 1024 * 1024) return $"{n / 1024.0 / 1024:0.#} MB";
        return $"{n / 1024.0 / 1024 / 1024:0.##} GB";
    }

    static string RelTime(string? iso)
    {
        if (string.IsNullOrWhiteSpace(iso) || iso.StartsWith("0001")) return "";
        if (!DateTimeOffset.TryParse(iso, out var t)) return iso;
        var d = DateTimeOffset.Now - t;
        if (d.TotalMinutes < 1) return "刚刚";
        if (d.TotalHours < 1) return $"{(int)d.TotalMinutes} 分钟前";
        if (d.TotalDays < 1) return t.ToLocalTime().ToString("HH:mm");
        if (d.TotalDays < 7) return t.ToLocalTime().ToString("M/d HH:mm");
        return t.ToLocalTime().ToString("yyyy/M/d");
    }

    static string Clock(string? iso)
    {
        if (string.IsNullOrWhiteSpace(iso)) return "";
        return DateTimeOffset.TryParse(iso, out var t) ? t.ToLocalTime().ToString("HH:mm") : iso;
    }
}

public sealed class HaloException(string message, int status) : Exception(message)
{
    public int Status { get; } = status;
}

static class Demo
{
    public static HostPerf Host() => new()
    {
        CpuModel = "Intel N100",
        Board = "GMKtec NucBox",
        Cpu = 18,
        Memory = 42,
        Disk = 61,
        Rx = 128_000,
        Tx = 54_000,
        MemoryText = "3.3 GB / 8 GB"
    };

    public static List<DeviceItem> Devices() =>
    [
        new() { Id = "mod-1", Name = "EC25 · 主卡", Operator = "中国联通", Mode = "LTE", Band = "B3", Imei = "860000000000001", Iccid = "8986…2201", PublicIp = "114.247.12.8", Phone = "+8613800138000", SignalDbm = -91, Healthy = true, Vowifi = true, ImsReady = true, Running = true, Radio = "vowifi" },
        new() { Id = "mod-2", Name = "EG25 · 备卡", Operator = "中国电信", Mode = "NR/LTE", Band = "n78", Imei = "860000000000002", Iccid = "8986…8830", PublicIp = "106.13.44.2", Phone = "+8613300133000", SignalDbm = -102, Healthy = true, Running = true, Radio = "cellular" },
        new() { Id = "mod-3", Name = "410 Stick", Operator = "Vodafone UK", Mode = "LTE", Band = "B20", PublicIp = "82.132.10.4", Phone = "+447700900123", SignalDbm = -87, Healthy = true, Vowifi = true, ImsReady = true, Running = true, Radio = "vowifi" }
    ];

    public static List<TaskRow> Tasks() =>
    [
        new() { Name = "夜间切卡", Next = "02:10" },
        new() { Name = "信号巡检", Next = "每 15 分钟" }
    ];

    public static List<SmsThread> Threads() =>
    [
        new() { Key = "1|+447700900000", Peer = "+447700900000", Display = "Vodafone", Preview = "Your code is 482193. Don't share it.", When = "2 分钟前", DeviceId = "mod-3", Unread = 1 },
        new() { Key = "1|+8613800138000", Peer = "+8613800138000", Display = "联通 10010", Preview = "您的流量已使用 2.4 GB", When = "昨天", DeviceId = "mod-1" },
        new() { Key = "1|+12025550123", Peer = "+12025550123", Display = "Alex", Preview = "Landed. Call you on WiFi Calling.", When = "周一", DeviceId = "mod-3" }
    ];

    public static List<SmsBubble> Bubbles(string peer)
    {
        if (peer.Contains("4477"))
            return
            [
                new() { Body = "Welcome to Vodafone. Your eSIM is active.", Time = "09:12", Mine = false },
                new() { Body = "Your code is 482193. Don't share it.", Time = "09:14", Mine = false }
            ];
        return
        [
            new() { Body = "信号稳定了，WiFi Calling 已注册。", Time = "18:02", Mine = false },
            new() { Body = "收到，今晚切到备卡。", Time = "18:04", Mine = true, Status = "已送达" },
            new() { Body = "Landed. Call you on WiFi Calling.", Time = "21:40", Mine = false }
        ];
    }

    public static List<CallRow> Calls() =>
    [
        new() { Number = "+447700900123", Direction = "outgoing", State = "answered", When = "今天 10:24", Duration = "4:12", Transport = "vowifi" },
        new() { Number = "+8613800138000", Direction = "incoming", State = "missed", When = "昨天 22:01", Duration = "—", Transport = "vowifi" },
        new() { Number = "+12025550123", Direction = "outgoing", State = "ended", When = "周一 19:18", Duration = "12:08", Transport = "vowifi" }
    ];

    public static List<LogRow> Logs() =>
    [
        new() { Time = "10:24:01", Level = "info", Message = "IMS registered on Vodafone UK ePDG" },
        new() { Time = "10:24:08", Level = "info", Message = "SMS MO submitted via IMS, 1/1 parts accepted" },
        new() { Time = "10:25:11", Level = "warn", Message = "RSRP dropped to -112 dBm on EG25" },
        new() { Time = "10:26:02", Level = "debug", Message = "QMI WDS start-network cid=17" }
    ];

    public static List<TunnelRow> Tunnels() =>
    [
        new() { Id = "wg0", Name = "home-core", Running = true, Endpoint = "vpn.example.net:51820" },
        new() { Id = "wg1", Name = "exit-uk", Running = false, Endpoint = "uk-1.edge:51820" }
    ];
}
