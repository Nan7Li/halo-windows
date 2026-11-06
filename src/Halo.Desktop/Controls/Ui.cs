using System.Windows;
using System.Windows.Controls;

namespace Halo.Desktop.Controls;

/// <summary>Attached properties shared by the control templates in Themes/Controls.xaml.</summary>
public static class Ui
{
    public static readonly DependencyProperty PlaceholderProperty = DependencyProperty.RegisterAttached(
        "Placeholder", typeof(string), typeof(Ui), new FrameworkPropertyMetadata(null));
    public static string? GetPlaceholder(DependencyObject o) => (string?)o.GetValue(PlaceholderProperty);
    public static void SetPlaceholder(DependencyObject o, string? v) => o.SetValue(PlaceholderProperty, v);

    /// <summary>Second line of text for control tiles (e.g. "已开启").</summary>
    public static readonly DependencyProperty DetailProperty = DependencyProperty.RegisterAttached(
        "Detail", typeof(string), typeof(Ui), new FrameworkPropertyMetadata(null));
    public static string? GetDetail(DependencyObject o) => (string?)o.GetValue(DetailProperty);
    public static void SetDetail(DependencyObject o, string? v) => o.SetValue(DetailProperty, v);

    /// <summary>Leading symbol for text fields and menu items.</summary>
    public static readonly DependencyProperty GlyphProperty = DependencyProperty.RegisterAttached(
        "Glyph", typeof(string), typeof(Ui), new FrameworkPropertyMetadata(null));
    public static string? GetGlyph(DependencyObject o) => (string?)o.GetValue(GlyphProperty);
    public static void SetGlyph(DependencyObject o, string? v) => o.SetValue(GlyphProperty, v);

    public static readonly DependencyProperty CornerRadiusProperty = DependencyProperty.RegisterAttached(
        "CornerRadius", typeof(CornerRadius), typeof(Ui), new FrameworkPropertyMetadata(new CornerRadius(10)));
    public static CornerRadius GetCornerRadius(DependencyObject o) => (CornerRadius)o.GetValue(CornerRadiusProperty);
    public static void SetCornerRadius(DependencyObject o, CornerRadius v) => o.SetValue(CornerRadiusProperty, v);

    /// <summary>Inline form fields (inside a list row) show no focus ring, like System Settings.</summary>
    public static readonly DependencyProperty NoFocusRingProperty = DependencyProperty.RegisterAttached(
        "NoFocusRing", typeof(bool), typeof(Ui), new FrameworkPropertyMetadata(false));
    public static bool GetNoFocusRing(DependencyObject o) => (bool)o.GetValue(NoFocusRingProperty);
    public static void SetNoFocusRing(DependencyObject o, bool v) => o.SetValue(NoFocusRingProperty, v);

    /// <summary>Red text / red highlight for destructive menu items and buttons.</summary>
    public static readonly DependencyProperty IsDestructiveProperty = DependencyProperty.RegisterAttached(
        "IsDestructive", typeof(bool), typeof(Ui), new FrameworkPropertyMetadata(false));
    public static bool GetIsDestructive(DependencyObject o) => (bool)o.GetValue(IsDestructiveProperty);
    public static void SetIsDestructive(DependencyObject o, bool v) => o.SetValue(IsDestructiveProperty, v);

    /// <summary>PasswordBox has no bindable Text, so the placeholder watches this flag instead.</summary>
    public static readonly DependencyProperty IsEmptyProperty = DependencyProperty.RegisterAttached(
        "IsEmpty", typeof(bool), typeof(Ui), new FrameworkPropertyMetadata(true));
    public static bool GetIsEmpty(DependencyObject o) => (bool)o.GetValue(IsEmptyProperty);
    public static void SetIsEmpty(DependencyObject o, bool v) => o.SetValue(IsEmptyProperty, v);

    public static readonly DependencyProperty WatchPasswordProperty = DependencyProperty.RegisterAttached(
        "WatchPassword", typeof(bool), typeof(Ui), new FrameworkPropertyMetadata(false, OnWatchPassword));
    public static bool GetWatchPassword(DependencyObject o) => (bool)o.GetValue(WatchPasswordProperty);
    public static void SetWatchPassword(DependencyObject o, bool v) => o.SetValue(WatchPasswordProperty, v);

    static void OnWatchPassword(DependencyObject d, DependencyPropertyChangedEventArgs e)
    {
        if (d is not PasswordBox pb) return;
        pb.PasswordChanged -= Changed;
        if ((bool)e.NewValue) pb.PasswordChanged += Changed;
        SetIsEmpty(pb, pb.Password.Length == 0);
        static void Changed(object s, RoutedEventArgs _) { var p = (PasswordBox)s; SetIsEmpty(p, p.Password.Length == 0); }
    }
}
