using System.Windows;
using System.Windows.Controls;

namespace Halo.Desktop.Controls;

/// <summary>
/// Widget grid: as many equal columns as fit at MinColumnWidth, items flow left→right and
/// may span columns (AdaptiveGrid.ColumnSpan). Rows take the tallest item's height.
/// </summary>
public class AdaptiveGrid : Panel
{
    public static readonly DependencyProperty MinColumnWidthProperty = DependencyProperty.Register(
        nameof(MinColumnWidth), typeof(double), typeof(AdaptiveGrid),
        new FrameworkPropertyMetadata(220.0, FrameworkPropertyMetadataOptions.AffectsMeasure));
    public static readonly DependencyProperty SpacingProperty = DependencyProperty.Register(
        nameof(Spacing), typeof(double), typeof(AdaptiveGrid),
        new FrameworkPropertyMetadata(12.0, FrameworkPropertyMetadataOptions.AffectsMeasure));
    public static readonly DependencyProperty ColumnSpanProperty = DependencyProperty.RegisterAttached(
        "ColumnSpan", typeof(int), typeof(AdaptiveGrid),
        new FrameworkPropertyMetadata(1, FrameworkPropertyMetadataOptions.AffectsParentMeasure));

    public double MinColumnWidth { get => (double)GetValue(MinColumnWidthProperty); set => SetValue(MinColumnWidthProperty, value); }
    public double Spacing { get => (double)GetValue(SpacingProperty); set => SetValue(SpacingProperty, value); }
    public static int GetColumnSpan(UIElement e) => (int)e.GetValue(ColumnSpanProperty);
    public static void SetColumnSpan(UIElement e, int v) => e.SetValue(ColumnSpanProperty, v);

    readonly List<(UIElement el, int row, int col, int span)> _layout = [];
    readonly List<double> _rowHeights = [];
    int _cols = 1;
    double _colW;

    protected override Size MeasureOverride(Size available)
    {
        var width = double.IsInfinity(available.Width) ? MinColumnWidth * 3 + Spacing * 2 : available.Width;
        var sp = Spacing;
        _cols = Math.Max(1, (int)Math.Floor((width + sp) / (MinColumnWidth + sp)));
        _colW = (width - sp * (_cols - 1)) / _cols;
        _layout.Clear();
        _rowHeights.Clear();

        int row = 0, col = 0;
        foreach (UIElement el in InternalChildren)
        {
            if (el.Visibility == Visibility.Collapsed) continue;
            var span = Math.Clamp(GetColumnSpan(el), 1, _cols);
            if (col + span > _cols) { row++; col = 0; }
            el.Measure(new Size(_colW * span + sp * (span - 1), double.PositiveInfinity));
            while (_rowHeights.Count <= row) _rowHeights.Add(0);
            _rowHeights[row] = Math.Max(_rowHeights[row], el.DesiredSize.Height);
            _layout.Add((el, row, col, span));
            col += span;
            if (col >= _cols) { row++; col = 0; }
        }
        var h = _rowHeights.Sum() + sp * Math.Max(0, _rowHeights.Count - 1);
        return new Size(width, h);
    }

    protected override Size ArrangeOverride(Size final)
    {
        var sp = Spacing;
        var tops = new double[_rowHeights.Count];
        for (var i = 1; i < tops.Length; i++) tops[i] = tops[i - 1] + _rowHeights[i - 1] + sp;
        foreach (var (el, row, col, span) in _layout)
            el.Arrange(new Rect(col * (_colW + sp), tops[row], _colW * span + sp * (span - 1), _rowHeights[row]));
        return final;
    }
}
