using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using System.Windows.Media;
using Halo.Desktop.Controls;

namespace Halo.Desktop.Services;

/// <summary>
/// Small modal sheets (text prompt, confirmation, alert) drawn with the app's own styles:
/// 18pt continuous-corner dialog, title + message, capsule buttons right-aligned with the
/// default action last (macOS order).
/// </summary>
static class Prompt
{
    public static string? Ask(Window? owner, string title, string hint, string seed = "")
    {
        var box = new TextBox { Text = seed, Margin = new Thickness(0, 12, 0, 0), Height = 28 };
        var ok = Show(owner, title, hint, box, "确定", destructive: false, cancel: "取消", focus: box);
        var text = box.Text.Trim();
        return ok && text.Length > 0 ? text : null;
    }

    public static bool Confirm(Window? owner, string title, string message, string action, bool destructive = false) =>
        Show(owner, title, message, null, action, destructive, cancel: "取消", focus: null);

    public static void Alert(Window? owner, string title, string message) =>
        Show(owner, title, message, null, "好", destructive: false, cancel: null, focus: null);

    static bool Show(Window? owner, string title, string message, UIElement? body, string action, bool destructive, string? cancel, UIElement? focus)
    {
        var res = Application.Current.Resources;
        var ok = false;
        var win = new Window
        {
            Title = title,
            Width = 340,
            SizeToContent = SizeToContent.Height,
            WindowStartupLocation = owner != null ? WindowStartupLocation.CenterOwner : WindowStartupLocation.CenterScreen,
            Owner = owner,
            WindowStyle = WindowStyle.None,
            AllowsTransparency = true,
            Background = Brushes.Transparent,
            ResizeMode = ResizeMode.NoResize,
            ShowInTaskbar = false,
            UseLayoutRounding = true,
            FontFamily = (FontFamily)res["UiFont"],
            FontSize = (double)res["FS.Body"],
        };
        win.SetResourceReference(Window.ForegroundProperty, "LabelBrush");
        win.SetResourceReference(Window.FontWeightProperty, "W.Regular");
        TextOptions.SetTextFormattingMode(win, TextFormattingMode.Display);

        var panel = new StackPanel();
        var titleBlock = new TextBlock { Text = title, TextWrapping = TextWrapping.Wrap };
        titleBlock.SetResourceReference(FrameworkElement.StyleProperty, "Text.Headline");
        panel.Children.Add(titleBlock);
        if (!string.IsNullOrWhiteSpace(message))
        {
            var msg = new TextBlock { Text = message, TextWrapping = TextWrapping.Wrap, Margin = new Thickness(0, 4, 0, 0) };
            msg.SetResourceReference(FrameworkElement.StyleProperty, "Text.Secondary");
            panel.Children.Add(msg);
        }
        if (body != null) panel.Children.Add(body);

        var row = new StackPanel { Orientation = Orientation.Horizontal, HorizontalAlignment = HorizontalAlignment.Right, Margin = new Thickness(0, 18, 0, 0) };
        if (cancel != null)
        {
            var c = new Button { Content = cancel, Margin = new Thickness(0, 0, 8, 0), MinWidth = 72, IsCancel = true };
            c.Click += (_, _) => win.Close();
            row.Children.Add(c);
        }
        var go = new Button { Content = action, MinWidth = 72, IsDefault = true };
        go.SetResourceReference(FrameworkElement.StyleProperty, destructive ? "Btn.Destructive" : "Btn.Primary");
        go.Click += (_, _) => { ok = true; win.Close(); };
        row.Children.Add(go);
        panel.Children.Add(row);

        var card = new SquircleBorder { CornerRadius = (double)res["D.Dialog"], Padding = new Thickness(20, 18, 20, 16), Child = panel, StrokeThickness = 1 };
        card.SetResourceReference(SquircleBorder.BackgroundProperty, "Grouped2Brush");
        card.SetResourceReference(SquircleBorder.BorderBrushProperty, "SeparatorBrush");
        var shadow = new SquircleBorder
        {
            CornerRadius = (double)res["D.Dialog"],
            Effect = new System.Windows.Media.Effects.DropShadowEffect { BlurRadius = 30, ShadowDepth = 8, Direction = 270, Opacity = 0.25, Color = Colors.Black }
        };
        shadow.SetResourceReference(SquircleBorder.BackgroundProperty, "Grouped2Brush");
        win.Content = new Grid { Margin = new Thickness(24), Children = { shadow, card } };
        win.MouseLeftButtonDown += (_, e) => { if (e.ButtonState == MouseButtonState.Pressed) try { win.DragMove(); } catch { } };
        win.Loaded += (_, _) => { if (focus is TextBox tb) { tb.Focus(); tb.CaretIndex = tb.Text.Length; } else go.Focus(); };
        win.ShowDialog();
        return ok;
    }
}
