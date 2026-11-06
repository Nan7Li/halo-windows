using System.Windows;
using System.Windows.Media;

namespace Halo.Desktop.Controls;

/// <summary>
/// One-device-pixel separator: 0.5 DIP at 200 % scaling, 1 DIP at 100 %. That is the macOS
/// hairline, and it never rounds away to zero the way a literal 0.5 height would at 100 %.
/// </summary>
public class Hairline : FrameworkElement
{
    public static readonly DependencyProperty StrokeProperty = DependencyProperty.Register(
        nameof(Stroke), typeof(Brush), typeof(Hairline),
        new FrameworkPropertyMetadata(null, FrameworkPropertyMetadataOptions.AffectsRender));

    public static readonly DependencyProperty VerticalProperty = DependencyProperty.Register(
        nameof(Vertical), typeof(bool), typeof(Hairline),
        new FrameworkPropertyMetadata(false, FrameworkPropertyMetadataOptions.AffectsMeasure));

    public Brush? Stroke { get => (Brush?)GetValue(StrokeProperty); set => SetValue(StrokeProperty, value); }
    public bool Vertical { get => (bool)GetValue(VerticalProperty); set => SetValue(VerticalProperty, value); }

    public Hairline()
    {
        SetResourceReference(StrokeProperty, "SeparatorBrush");
        IsHitTestVisible = false;
        SnapsToDevicePixels = true;
    }

    /// <summary>Screenshot harness only: render as if the monitor had this scale factor.</summary>
    internal static double? ScaleOverride { get; set; }

    double Px => 1 / (ScaleOverride ?? (Vertical ? VisualTreeHelper.GetDpi(this).DpiScaleX : VisualTreeHelper.GetDpi(this).DpiScaleY));

    protected override Size MeasureOverride(Size availableSize) =>
        Vertical ? new Size(Px, 0) : new Size(0, Px);

    protected override void OnDpiChanged(DpiScale oldDpi, DpiScale newDpi) => InvalidateMeasure();

    protected override void OnRender(DrawingContext dc)
    {
        if (Stroke == null) return;
        var px = Px;
        var rect = Vertical ? new Rect(0, 0, px, ActualHeight) : new Rect(0, 0, ActualWidth, px);
        var guides = new GuidelineSet([rect.Left, rect.Right], [rect.Top, rect.Bottom]);
        dc.PushGuidelineSet(guides);
        dc.DrawRectangle(Stroke, null, rect);
        dc.Pop();
    }
}
