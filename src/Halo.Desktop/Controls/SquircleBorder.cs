using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;

namespace Halo.Desktop.Controls;

/// <summary>
/// Border replacement that draws a continuous-corner shape and clips its child to it.
/// Used for every large surface (grouped lists, cards, glass panels, dialogs, menus).
/// </summary>
public class SquircleBorder : Decorator
{
    public static readonly DependencyProperty CornerRadiusProperty = DependencyProperty.Register(
        nameof(CornerRadius), typeof(double), typeof(SquircleBorder),
        new FrameworkPropertyMetadata(14.0, FrameworkPropertyMetadataOptions.AffectsRender | FrameworkPropertyMetadataOptions.AffectsArrange));

    public static readonly DependencyProperty SmoothingProperty = DependencyProperty.Register(
        nameof(Smoothing), typeof(double), typeof(SquircleBorder),
        new FrameworkPropertyMetadata(0.6, FrameworkPropertyMetadataOptions.AffectsRender | FrameworkPropertyMetadataOptions.AffectsArrange));

    public static readonly DependencyProperty BackgroundProperty = Panel.BackgroundProperty.AddOwner(
        typeof(SquircleBorder), new FrameworkPropertyMetadata(null, FrameworkPropertyMetadataOptions.AffectsRender));

    public static readonly DependencyProperty BorderBrushProperty = Border.BorderBrushProperty.AddOwner(
        typeof(SquircleBorder), new FrameworkPropertyMetadata(null, FrameworkPropertyMetadataOptions.AffectsRender));

    public static readonly DependencyProperty StrokeThicknessProperty = DependencyProperty.Register(
        nameof(StrokeThickness), typeof(double), typeof(SquircleBorder),
        new FrameworkPropertyMetadata(0.0, FrameworkPropertyMetadataOptions.AffectsRender));

    public static readonly DependencyProperty PaddingProperty = Border.PaddingProperty.AddOwner(
        typeof(SquircleBorder), new FrameworkPropertyMetadata(new Thickness(), FrameworkPropertyMetadataOptions.AffectsMeasure));

    public static readonly DependencyProperty ClipContentProperty = DependencyProperty.Register(
        nameof(ClipContent), typeof(bool), typeof(SquircleBorder),
        new FrameworkPropertyMetadata(true, FrameworkPropertyMetadataOptions.AffectsArrange));

    public double CornerRadius { get => (double)GetValue(CornerRadiusProperty); set => SetValue(CornerRadiusProperty, value); }
    public double Smoothing { get => (double)GetValue(SmoothingProperty); set => SetValue(SmoothingProperty, value); }
    public Brush? Background { get => (Brush?)GetValue(BackgroundProperty); set => SetValue(BackgroundProperty, value); }
    public Brush? BorderBrush { get => (Brush?)GetValue(BorderBrushProperty); set => SetValue(BorderBrushProperty, value); }
    public double StrokeThickness { get => (double)GetValue(StrokeThicknessProperty); set => SetValue(StrokeThicknessProperty, value); }
    public Thickness Padding { get => (Thickness)GetValue(PaddingProperty); set => SetValue(PaddingProperty, value); }
    public bool ClipContent { get => (bool)GetValue(ClipContentProperty); set => SetValue(ClipContentProperty, value); }

    Geometry? _shape;
    Size _shapeSize;

    protected override Size MeasureOverride(Size constraint)
    {
        var pad = Padding;
        var ph = pad.Left + pad.Right;
        var pv = pad.Top + pad.Bottom;
        if (Child == null) return new Size(ph, pv);
        Child.Measure(new Size(Math.Max(0, constraint.Width - ph), Math.Max(0, constraint.Height - pv)));
        return new Size(Child.DesiredSize.Width + ph, Child.DesiredSize.Height + pv);
    }

    protected override Size ArrangeOverride(Size arrangeSize)
    {
        var pad = Padding;
        Child?.Arrange(new Rect(pad.Left, pad.Top,
            Math.Max(0, arrangeSize.Width - pad.Left - pad.Right),
            Math.Max(0, arrangeSize.Height - pad.Top - pad.Bottom)));
        _shape = null;
        Clip = ClipContent ? Shape(arrangeSize) : null;
        return arrangeSize;
    }

    Geometry Shape(Size size)
    {
        if (_shape == null || _shapeSize != size)
        {
            _shape = Squircle.Create(new Rect(size), CornerRadius, Smoothing);
            _shapeSize = size;
        }
        return _shape;
    }

    protected override void OnRender(DrawingContext dc)
    {
        var size = RenderSize;
        if (size.Width <= 0 || size.Height <= 0) return;
        if (Background != null) dc.DrawGeometry(Background, null, Shape(size));
        var t = StrokeThickness;
        if (BorderBrush != null && t > 0)
        {
            var inner = Squircle.Create(new Rect(t / 2, t / 2, Math.Max(0, size.Width - t), Math.Max(0, size.Height - t)),
                Math.Max(0, CornerRadius - t / 2), Smoothing);
            dc.DrawGeometry(null, new Pen(BorderBrush, t), inner);
        }
    }
}
