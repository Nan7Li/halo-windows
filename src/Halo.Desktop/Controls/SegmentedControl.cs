using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using Halo.Desktop.Services;

namespace Halo.Desktop.Controls;

/// <summary>
/// macOS/iOS segmented control: fill3 capsule track, raised white (grouped) capsule under
/// the selected segment that slides on a spring. Items are plain strings.
/// </summary>
[TemplatePart(Name = "PART_Indicator", Type = typeof(FrameworkElement))]
public class SegmentedControl : ListBox
{
    const double Inset = 2;
    FrameworkElement? _indicator;
    readonly TranslateTransform _tt = new();
    bool _placed;

    public SegmentedControl()
    {
        SelectionMode = SelectionMode.Single;
        SizeChanged += (_, _) => Place(false);
    }

    public override void OnApplyTemplate()
    {
        base.OnApplyTemplate();
        _indicator = GetTemplateChild("PART_Indicator") as FrameworkElement;
        if (_indicator != null) _indicator.RenderTransform = _tt;
        Place(false);
    }

    protected override void OnSelectionChanged(SelectionChangedEventArgs e)
    {
        base.OnSelectionChanged(e);
        Place(_placed);
    }

    void Place(bool animate)
    {
        if (_indicator == null || Items.Count == 0 || ActualWidth <= 0) return;
        var seg = (ActualWidth - Inset * 2) / Items.Count;
        _indicator.Width = Math.Max(0, seg);
        _indicator.Visibility = SelectedIndex < 0 ? Visibility.Hidden : Visibility.Visible;
        var x = Math.Max(0, SelectedIndex) * seg;
        if (animate)
        {
            if (Motion.Level == MotionLevel.Reduced)
            {
                // No travel: jump, then fade the thumb back in.
                Motion.Slide(_tt, TranslateTransform.XProperty, x);
                _indicator.Opacity = 0;
                Motion.Fade(_indicator, 1, 150);
            }
            else Motion.Slide(_tt, TranslateTransform.XProperty, x);
        }
        else
        {
            _tt.BeginAnimation(TranslateTransform.XProperty, null);
            _tt.X = x;
        }
        _placed = true;
    }
}
