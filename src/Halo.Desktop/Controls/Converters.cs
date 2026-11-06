using System.Globalization;
using System.Windows;
using System.Windows.Data;
using Halo.Desktop.Models;

namespace Halo.Desktop.Controls;

/// <summary>
/// UI-side interpretation of model values. Models/ keeps its own legacy brushes; the new
/// templates ignore those and map state to the semantic palette through these converters.
/// </summary>
public static class CallInfo
{
    public static bool IsMissed(CallRow c)
    {
        var s = (c.State ?? "").ToLowerInvariant();
        return s.Contains("miss") || s.Contains("未接") || s.Contains("no_answer") || s.Contains("noanswer");
    }

    public static bool IsIncoming(CallRow c) => string.Equals(c.Direction, "incoming", StringComparison.OrdinalIgnoreCase);

    public static string Describe(CallRow c)
    {
        var dir = IsMissed(c) ? "未接来电" : IsIncoming(c) ? "来电" : "呼出";
        var transport = string.IsNullOrWhiteSpace(c.Transport) ? "" : " · " + (c.Transport.Equals("vowifi", StringComparison.OrdinalIgnoreCase) ? "WiFi Calling" : c.Transport);
        return $"{dir}{transport}";
    }
}

public sealed class MissedCallConverter : IValueConverter
{
    public object Convert(object value, Type t, object p, CultureInfo c) => value is CallRow r && CallInfo.IsMissed(r);
    public object ConvertBack(object v, Type t, object p, CultureInfo c) => Binding.DoNothing;
}

public sealed class CallGlyphConverter : IValueConverter
{
    public object Convert(object value, Type t, object p, CultureInfo c) =>
        value is CallRow r && CallInfo.IsIncoming(r) ? Glyphs.Incoming : Glyphs.Outgoing;
    public object ConvertBack(object v, Type t, object p, CultureInfo c) => Binding.DoNothing;
}

public sealed class CallDescriptionConverter : IValueConverter
{
    public object Convert(object value, Type t, object p, CultureInfo c) => value is CallRow r ? CallInfo.Describe(r) : "";
    public object ConvertBack(object v, Type t, object p, CultureInfo c) => Binding.DoNothing;
}

/// <summary>Normalises a log level to error / warn / debug / info.</summary>
public sealed class LogKindConverter : IValueConverter
{
    public static string Kind(string? level) => (level ?? "").ToLowerInvariant() switch
    {
        "error" or "fatal" or "err" or "crit" or "critical" => "error",
        "warn" or "warning" => "warn",
        "debug" or "trace" => "debug",
        _ => "info"
    };
    public object Convert(object value, Type t, object p, CultureInfo c) => Kind(value as string);
    public object ConvertBack(object v, Type t, object p, CultureInfo c) => Binding.DoNothing;
}

public sealed class UpperConverter : IValueConverter
{
    public object Convert(object value, Type t, object p, CultureInfo c) => (value as string ?? "").ToUpperInvariant();
    public object ConvertBack(object v, Type t, object p, CultureInfo c) => Binding.DoNothing;
}

/// <summary>Collapses an element when the bound string is null, empty or "—".</summary>
public sealed class EmptyToCollapsedConverter : IValueConverter
{
    public object Convert(object value, Type t, object p, CultureInfo c) =>
        value is string s && s.Length > 0 && s != "—" ? Visibility.Visible : Visibility.Collapsed;
    public object ConvertBack(object v, Type t, object p, CultureInfo c) => Binding.DoNothing;
}
