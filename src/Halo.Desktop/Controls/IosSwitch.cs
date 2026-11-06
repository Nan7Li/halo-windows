using System.Windows;
using System.Windows.Controls;
using System.Windows.Controls.Primitives;
using System.Windows.Media;
using Halo.Desktop.Services;

namespace Halo.Desktop.Controls;

/// <summary>
/// iOS 26 switch: 63×28 track, 38×24 capsule thumb. On = green, off = fill3.
/// Thumb travels on a spring (Full motion); Reduced keeps only the colour fade; Off is instant.
/// </summary>
[TemplatePart(Name = "PART_Thumb", Type = typeof(FrameworkElement))]
[TemplatePart(Name = "PART_On", Type = typeof(FrameworkElement))]
public class IosSwitch : ToggleButton
{
    public const double TrackWidth = 63, TrackHeight = 28, ThumbWidth = 38, ThumbHeight = 24, Inset = 2;
    static double Travel => TrackWidth - ThumbWidth - Inset * 2;

    FrameworkElement? _thumb, _on;
    readonly TranslateTransform _tt = new();

    public override void OnApplyTemplate()
    {
        base.OnApplyTemplate();
        _thumb = GetTemplateChild("PART_Thumb") as FrameworkElement;
        _on = GetTemplateChild("PART_On") as FrameworkElement;
        if (_thumb != null) _thumb.RenderTransform = _tt;
        Sync(false);
    }

    protected override void OnChecked(RoutedEventArgs e) { base.OnChecked(e); Sync(true); }
    protected override void OnUnchecked(RoutedEventArgs e) { base.OnUnchecked(e); Sync(true); }

    void Sync(bool animate)
    {
        var on = IsChecked == true;
        var x = on ? Travel : 0;
        if (_on != null)
        {
            if (animate) Motion.Fade(_on, on ? 1 : 0, 200);
            else { _on.BeginAnimation(OpacityProperty, null); _on.Opacity = on ? 1 : 0; }
        }
        if (animate) Motion.Slide(_tt, TranslateTransform.XProperty, x);
        else { _tt.BeginAnimation(TranslateTransform.XProperty, null); _tt.X = x; }
    }
}
