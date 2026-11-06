using System.Windows;
using System.Windows.Media.Animation;

namespace Halo.Desktop.Controls;

/// <summary>
/// Under-damped spring, normalised to the animation's duration. Damping 0.72 gives about a
/// 4 % overshoot, the "slight bounce" used for switches, segmented thumbs and the sidebar
/// selection. Use with EasingMode.EaseIn (the default here) so the curve is applied as-is.
/// </summary>
public sealed class SpringEase : EasingFunctionBase
{
    public static readonly DependencyProperty DampingProperty = DependencyProperty.Register(
        nameof(Damping), typeof(double), typeof(SpringEase), new PropertyMetadata(0.72));

    public double Damping { get => (double)GetValue(DampingProperty); set => SetValue(DampingProperty, value); }

    public SpringEase() => EasingMode = EasingMode.EaseIn;

    protected override double EaseInCore(double t)
    {
        if (t >= 1) return 1;
        const double omega = 10.0;
        var z = Math.Clamp(Damping, 0.05, 0.99);
        var wd = omega * Math.Sqrt(1 - z * z);
        return 1 - Math.Exp(-z * omega * t) * (Math.Cos(wd * t) + z * omega / wd * Math.Sin(wd * t));
    }

    protected override Freezable CreateInstanceCore() => new SpringEase();
}
