using System.Windows;
using System.Windows.Controls;

namespace Halo.Desktop.Controls;

/// <summary>
/// Liquid-glass surface for the navigation/control layer only (sidebar, toolbar button
/// groups, login card). Translucent fill + 1px top-light/bottom-dark gradient rim + soft
/// shadow. Fill opacity follows the "玻璃质感" preference (GlassFillBrush). Template lives
/// in Themes/Controls.xaml.
/// </summary>
public class GlassPanel : ContentControl
{
    public static readonly DependencyProperty CornerRadiusProperty = DependencyProperty.Register(
        nameof(CornerRadius), typeof(double), typeof(GlassPanel), new PropertyMetadata(22.0));

    public static readonly DependencyProperty ElevatedProperty = DependencyProperty.Register(
        nameof(Elevated), typeof(bool), typeof(GlassPanel), new PropertyMetadata(true));

    public double CornerRadius { get => (double)GetValue(CornerRadiusProperty); set => SetValue(CornerRadiusProperty, value); }
    public bool Elevated { get => (bool)GetValue(ElevatedProperty); set => SetValue(ElevatedProperty, value); }
}
