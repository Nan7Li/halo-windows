using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;

namespace Halo.Desktop.Controls;

/// <summary>Blurred value (IMEI, ICCID …) that reveals on click and hides again on a second click.</summary>
public class SensitiveText : Control
{
    public static readonly DependencyProperty TextProperty = DependencyProperty.Register(
        nameof(Text), typeof(string), typeof(SensitiveText), new PropertyMetadata(""));
    public static readonly DependencyProperty IsRevealedProperty = DependencyProperty.Register(
        nameof(IsRevealed), typeof(bool), typeof(SensitiveText), new PropertyMetadata(false));

    public string Text { get => (string)GetValue(TextProperty); set => SetValue(TextProperty, value); }
    public bool IsRevealed { get => (bool)GetValue(IsRevealedProperty); set => SetValue(IsRevealedProperty, value); }

    public SensitiveText()
    {
        Cursor = Cursors.Hand;
        Focusable = true;
    }

    protected override void OnMouseLeftButtonUp(MouseButtonEventArgs e)
    {
        base.OnMouseLeftButtonUp(e);
        IsRevealed = !IsRevealed;
        e.Handled = true;
    }

    protected override void OnKeyDown(KeyEventArgs e)
    {
        base.OnKeyDown(e);
        if (e.Key is Key.Enter or Key.Space) { IsRevealed = !IsRevealed; e.Handled = true; }
    }
}
