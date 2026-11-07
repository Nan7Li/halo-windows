using System.Windows;
using System.Windows.Controls;

namespace Halo.Desktop.Controls;

/// <summary>
/// Empty state (HIG "ContentUnavailableView"): no frame, large label3 symbol,
/// a title and a one-line description, centred.
/// </summary>
public class ContentUnavailable : Control
{
    public static readonly DependencyProperty GlyphProperty = DependencyProperty.Register(
        nameof(Glyph), typeof(string), typeof(ContentUnavailable), new PropertyMetadata(Glyphs.Info));
    public static readonly DependencyProperty TitleProperty = DependencyProperty.Register(
        nameof(Title), typeof(string), typeof(ContentUnavailable), new PropertyMetadata(""));
    public static readonly DependencyProperty DescriptionProperty = DependencyProperty.Register(
        nameof(Description), typeof(string), typeof(ContentUnavailable), new PropertyMetadata(null));

    public string Glyph { get => (string)GetValue(GlyphProperty); set => SetValue(GlyphProperty, value); }
    public string Title { get => (string)GetValue(TitleProperty); set => SetValue(TitleProperty, value); }
    public string? Description { get => (string?)GetValue(DescriptionProperty); set => SetValue(DescriptionProperty, value); }
}
