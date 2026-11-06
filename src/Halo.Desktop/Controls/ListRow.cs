using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;
using System.Windows.Media;

namespace Halo.Desktop.Controls;

/// <summary>
/// One row of a <see cref="ListSection"/>: label on the left, value/control on the right
/// (the row's Content), 38pt minimum height. Draws a hairline separator above itself
/// unless it is the first row, inset to where the title text starts.
/// </summary>
[TemplatePart(Name = "PART_Text", Type = typeof(FrameworkElement))]
[TemplatePart(Name = "PART_Separator", Type = typeof(FrameworkElement))]
public class ListRow : ContentControl
{
    public static readonly DependencyProperty TitleProperty = DependencyProperty.Register(
        nameof(Title), typeof(string), typeof(ListRow), new PropertyMetadata(null));
    public static readonly DependencyProperty SubtitleProperty = DependencyProperty.Register(
        nameof(Subtitle), typeof(string), typeof(ListRow), new PropertyMetadata(null));
    public static readonly DependencyProperty GlyphProperty = DependencyProperty.Register(
        nameof(Glyph), typeof(string), typeof(ListRow), new PropertyMetadata(null));
    public static readonly DependencyProperty GlyphBackgroundProperty = DependencyProperty.Register(
        nameof(GlyphBackground), typeof(Brush), typeof(ListRow), new PropertyMetadata(null));
    public static readonly DependencyProperty GlyphForegroundProperty = DependencyProperty.Register(
        nameof(GlyphForeground), typeof(Brush), typeof(ListRow), new PropertyMetadata(Brushes.White));
    public static readonly DependencyProperty LeadingProperty = DependencyProperty.Register(
        nameof(Leading), typeof(object), typeof(ListRow), new PropertyMetadata(null));
    public static readonly DependencyProperty ShowChevronProperty = DependencyProperty.Register(
        nameof(ShowChevron), typeof(bool), typeof(ListRow), new PropertyMetadata(false));
    public static readonly DependencyProperty IsClickableProperty = DependencyProperty.Register(
        nameof(IsClickable), typeof(bool), typeof(ListRow), new PropertyMetadata(false));
    public static readonly DependencyProperty IsDestructiveProperty = DependencyProperty.Register(
        nameof(IsDestructive), typeof(bool), typeof(ListRow), new PropertyMetadata(false));
    public static readonly DependencyProperty ShowSeparatorProperty = DependencyProperty.Register(
        nameof(ShowSeparator), typeof(bool), typeof(ListRow), new PropertyMetadata(false));
    public static readonly DependencyProperty IsPressedProperty = DependencyProperty.Register(
        nameof(IsPressed), typeof(bool), typeof(ListRow), new PropertyMetadata(false));

    public static readonly RoutedEvent ClickEvent = EventManager.RegisterRoutedEvent(
        nameof(Click), RoutingStrategy.Bubble, typeof(RoutedEventHandler), typeof(ListRow));

    public string? Title { get => (string?)GetValue(TitleProperty); set => SetValue(TitleProperty, value); }
    public string? Subtitle { get => (string?)GetValue(SubtitleProperty); set => SetValue(SubtitleProperty, value); }
    public string? Glyph { get => (string?)GetValue(GlyphProperty); set => SetValue(GlyphProperty, value); }
    public Brush? GlyphBackground { get => (Brush?)GetValue(GlyphBackgroundProperty); set => SetValue(GlyphBackgroundProperty, value); }
    public Brush? GlyphForeground { get => (Brush?)GetValue(GlyphForegroundProperty); set => SetValue(GlyphForegroundProperty, value); }
    public object? Leading { get => GetValue(LeadingProperty); set => SetValue(LeadingProperty, value); }
    public bool ShowChevron { get => (bool)GetValue(ShowChevronProperty); set => SetValue(ShowChevronProperty, value); }
    public bool IsClickable { get => (bool)GetValue(IsClickableProperty); set => SetValue(IsClickableProperty, value); }
    public bool IsDestructive { get => (bool)GetValue(IsDestructiveProperty); set => SetValue(IsDestructiveProperty, value); }
    public bool ShowSeparator { get => (bool)GetValue(ShowSeparatorProperty); set => SetValue(ShowSeparatorProperty, value); }
    public bool IsPressed { get => (bool)GetValue(IsPressedProperty); private set => SetValue(IsPressedProperty, value); }

    public event RoutedEventHandler Click { add => AddHandler(ClickEvent, value); remove => RemoveHandler(ClickEvent, value); }

    FrameworkElement? _text, _sep;

    public ListRow()
    {
        Loaded += (_, _) => UpdateSeparator();
    }

    public override void OnApplyTemplate()
    {
        base.OnApplyTemplate();
        _text = GetTemplateChild("PART_Text") as FrameworkElement;
        _sep = GetTemplateChild("PART_Separator") as FrameworkElement;
    }

    protected override Size ArrangeOverride(Size arrangeBounds)
    {
        var r = base.ArrangeOverride(arrangeBounds);
        if (_text != null && _sep != null && _text.IsVisible)
        {
            var x = _text.TranslatePoint(new Point(0, 0), this).X;
            if (Math.Abs(_sep.Margin.Left - x) > 0.1) _sep.Margin = new Thickness(x, 0, 0, 0);
        }
        return r;
    }

    /// <summary>First row in its list gets no separator.</summary>
    public void UpdateSeparator()
    {
        DependencyObject element = this;
        var parent = VisualTreeHelper.GetParent(element);
        while (parent != null && parent is not Panel)
        {
            element = parent;
            parent = VisualTreeHelper.GetParent(element);
        }
        if (parent is not Panel panel) { ShowSeparator = false; return; }
        var index = 0;
        foreach (UIElement child in panel.Children)
        {
            if (ReferenceEquals(child, element)) break;
            if (child.Visibility == Visibility.Visible) index++;
        }
        ShowSeparator = index > 0;
    }

    protected override void OnMouseLeftButtonDown(MouseButtonEventArgs e)
    {
        base.OnMouseLeftButtonDown(e);
        if (!IsClickable) return;
        IsPressed = true;
        CaptureMouse();
        e.Handled = true;
    }

    protected override void OnMouseLeftButtonUp(MouseButtonEventArgs e)
    {
        base.OnMouseLeftButtonUp(e);
        if (!IsPressed) return;
        IsPressed = false;
        ReleaseMouseCapture();
        if (IsMouseOver) RaiseEvent(new RoutedEventArgs(ClickEvent, this));
        e.Handled = true;
    }

    protected override void OnLostMouseCapture(MouseEventArgs e)
    {
        base.OnLostMouseCapture(e);
        IsPressed = false;
    }

    protected override void OnKeyDown(KeyEventArgs e)
    {
        base.OnKeyDown(e);
        if (IsClickable && e.Key is Key.Enter or Key.Space)
        {
            RaiseEvent(new RoutedEventArgs(ClickEvent, this));
            e.Handled = true;
        }
    }
}
