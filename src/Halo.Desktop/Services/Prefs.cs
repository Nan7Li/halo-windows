using System.IO;
using System.Text.Json;

namespace Halo.Desktop.Services;

public sealed class Prefs
{
    public string Server { get; set; } = "http://127.0.0.1:7575";
    public string Username { get; set; } = "admin";
    public bool Remember { get; set; } = true;
    public bool Dark { get; set; } = true;

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
        var dir = System.IO.Path.GetDirectoryName(Path)!;
        Directory.CreateDirectory(dir);
        File.WriteAllText(Path, JsonSerializer.Serialize(this, new JsonSerializerOptions { WriteIndented = true }));
    }
}
