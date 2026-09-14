using System.Windows;
using System.Windows.Controls;

namespace Halo.Desktop.Services;

static class Prompt
{
    public static string? Ask(Window? owner, string title, string hint, string seed = "")
    {
        var box = new TextBox { Text = seed, Margin = new Thickness(0, 8, 0, 0) };
        var ok = false;
        var win = new Window
        {
            Title = title,
            Width = 380,
            Height = 168,
            WindowStartupLocation = WindowStartupLocation.CenterOwner,
            Owner = owner,
            WindowStyle = WindowStyle.None,
            AllowsTransparency = true,
            Background = System.Windows.Media.Brushes.Transparent,
            ResizeMode = ResizeMode.NoResize
        };
        var panel = new StackPanel { Margin = new Thickness(22) };
        panel.Children.Add(new TextBlock { Text = hint, Opacity = 0.7, FontSize = 12 });
        panel.Children.Add(box);
        var row = new StackPanel { Orientation = Orientation.Horizontal, HorizontalAlignment = HorizontalAlignment.Right, Margin = new Thickness(0, 16, 0, 0) };
        var cancel = new Button { Content = "取消", Margin = new Thickness(0, 0, 8, 0) };
        var go = new Button { Content = "确定", Style = (Style)Application.Current.FindResource("AccentBtn") };
        cancel.Click += (_, _) => win.Close();
        go.Click += (_, _) => { ok = true; win.Close(); };
        row.Children.Add(cancel);
        row.Children.Add(go);
        panel.Children.Add(row);
        var chrome = new Border
        {
            CornerRadius = new CornerRadius(18),
            Background = (System.Windows.Media.Brush)Application.Current.FindResource("Glass2Brush"),
            BorderBrush = (System.Windows.Media.Brush)Application.Current.FindResource("LineBrush"),
            BorderThickness = new Thickness(1),
            Child = panel
        };
        win.Content = chrome;
        win.ShowDialog();
        var text = box.Text.Trim();
        return ok && text.Length > 0 ? text : null;
    }
}
