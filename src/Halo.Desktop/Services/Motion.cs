using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Windows.Media.Animation;
using Halo.Desktop.Controls;

namespace Halo.Desktop.Services;

public enum MotionLevel { Full, Reduced, Off }

/// <summary>
/// Central place for every animation, so the "动态效果" preference applies everywhere:
/// Full = springs + fades, Reduced = fades only (no translation/scale/bounce), Off = instant.
/// </summary>
public static class Motion
{
    public static MotionLevel Level { get; set; } = MotionLevel.Full;

    public const int SpringMs = 320;
    public const int FadeMs = 180;

    /// <summary>Moves a translate transform. Only animates at Full; otherwise jumps.</summary>
    public static void Slide(TranslateTransform tt, DependencyProperty prop, double to, int ms = SpringMs)
    {
        if (Level != MotionLevel.Full)
        {
            tt.BeginAnimation(prop, null);
            tt.SetValue(prop, to);
            return;
        }
        tt.BeginAnimation(prop, new DoubleAnimation(to, TimeSpan.FromMilliseconds(ms)) { EasingFunction = new SpringEase() });
    }

    /// <summary>Animates any double property with the spring at Full, jumps otherwise.</summary>
    public static void Spring(IAnimatable target, DependencyProperty prop, double to, int ms = SpringMs)
    {
        if (Level != MotionLevel.Full)
        {
            target.BeginAnimation(prop, null);
            ((DependencyObject)target).SetValue(prop, to);
            return;
        }
        target.BeginAnimation(prop, new DoubleAnimation(to, TimeSpan.FromMilliseconds(ms)) { EasingFunction = new SpringEase() });
    }

    /// <summary>Opacity fade. Runs at Full and Reduced; instant when Off.</summary>
    public static void Fade(UIElement el, double to, int ms = FadeMs, Action? done = null)
    {
        if (Level == MotionLevel.Off)
        {
            el.BeginAnimation(UIElement.OpacityProperty, null);
            el.Opacity = to;
            done?.Invoke();
            return;
        }
        var a = new DoubleAnimation(to, TimeSpan.FromMilliseconds(ms)) { EasingFunction = new CubicEase { EasingMode = EasingMode.EaseOut } };
        if (done != null) a.Completed += (_, _) => done();
        el.BeginAnimation(UIElement.OpacityProperty, a);
    }

    /// <summary>Cross-fades two stacked pages (150–200 ms). The old page stops hit-testing at once.</summary>
    public static void CrossFade(UIElement? from, UIElement to)
    {
        var start = to.Visibility == Visibility.Visible ? to.Opacity : 0;
        to.BeginAnimation(UIElement.OpacityProperty, null);
        to.Opacity = start;
        to.Visibility = Visibility.Visible;
        to.IsHitTestVisible = true;
        if (from == null || ReferenceEquals(from, to))
        {
            to.Opacity = 1;
            return;
        }
        from.IsHitTestVisible = false;
        Panel.SetZIndex(to, 1);
        Panel.SetZIndex(from, 0);
        Fade(to, 1);
        Fade(from, 0, done: () =>
        {
            if (!from.IsHitTestVisible) from.Visibility = Visibility.Collapsed;
        });
    }
}
