using System.IO;
using System.Text.Json;

namespace Halo.Desktop.Services;

public sealed class Prefs
{
    public string Server { get; set; } = "http://127.0.0.1:7575";
    public string Username { get; set; } = "admin";
    public bool Remember { get; set; } = true;
    /// <summary>Kept so older prefs.json files still load; superseded by <see cref="Appearance"/>.</summary>
    public bool Dark { get; set; } = true;

    /// <summary>"system" (default, follows Windows), "light" or "dark".</summary>
    public string Appearance { get; set; } = "system";
    /// <summary>Accent key from <see cref="ThemeManager.Accents"/>; default system blue.</summary>
    public string Accent { get; set; } = "blue";
    /// <summary>Glass clarity 0 = clear … 1 = tinted.</summary>
    public double Glass { get; set; } = 0.5;
    /// <summary>"full", "reduced" (fades only) or "off".</summary>
    public string Motion { get; set; } = "full";
    public bool BoldText { get; set; }

    static string Path => System.IO.Path.Combine(
        Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
        "HaloDesktop", "prefs.json");

    public static Prefs Load()
    {
        try
        {
            if (File.Exists(Path))
                return JsonSerializer.Deserialize<Prefs>(File.ReadAllText(Path)) ?? new Prefs();
        }
        catch { /* first launch */ }
        return new Prefs();
    }

    public void Save()
    {
        try
        {
            var dir = System.IO.Path.GetDirectoryName(Path)!;
            Directory.CreateDirectory(dir);
            File.WriteAllText(Path, JsonSerializer.Serialize(this, new JsonSerializerOptions { WriteIndented = true }));
        }
        catch { /* prefs are best-effort */ }
    }
}
