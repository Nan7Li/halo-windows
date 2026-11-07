using System.Windows;
using System.Windows.Media;

namespace Halo.Desktop.Controls;

/// <summary>
/// Continuous-curvature rounded rectangle ("squircle"), the corner shape Apple uses
/// instead of plain circular arcs. Port of Figma's corner-smoothing construction:
/// each corner is bezier → short arc → bezier, so curvature ramps up instead of
/// jumping at the tangent point. smoothing = 0.6 matches the iOS look.
/// </summary>
public static class Squircle
{
    public static Geometry Create(Rect r, double radius, double smoothing = 0.6)
    {
        if (r.Width <= 0 || r.Height <= 0) return Geometry.Empty;
        var budget = Math.Min(r.Width, r.Height) / 2;
        var cr = Math.Min(radius, budget);
        if (cr <= 0.01)
        {
            var rect = new RectangleGeometry(r);
            rect.Freeze();
            return rect;
        }

        var p = Math.Min((1 + smoothing) * cr, budget);
        var s = Math.Min(smoothing, Math.Max(p / cr - 1, 0));
        var arcMeasure = 90 * (1 - s);
        var arc = Math.Sin(Rad(arcMeasure / 2)) * cr * Math.Sqrt(2);
        var alpha = (90 - arcMeasure) / 2;
        var p3p4 = cr * Math.Tan(Rad(alpha / 2));
        var beta = 45 * s;
        var c = p3p4 * Math.Cos(Rad(beta));
        var d = c * Math.Tan(Rad(beta));
        var b = (p - arc - c - d) / 3;
        var a = 2 * b;

        double w = r.Width, h = r.Height, x0 = r.X, y0 = r.Y;
        var g = new StreamGeometry();
        using (var ctx = g.Open())
        {
            var cur = new Point(x0 + w - p, y0);
            ctx.BeginFigure(cur, true, true);

            // top-right
            Bez(ctx, ref cur, a, 0, a + b, 0, a + b + c, d);
            Arc(ctx, ref cur, arc, arc, cr);
            Bez(ctx, ref cur, d, c, d, b + c, d, a + b + c);
            Line(ctx, ref cur, new Point(x0 + w, y0 + h - p));
            // bottom-right
            Bez(ctx, ref cur, 0, a, 0, a + b, -d, a + b + c);
            Arc(ctx, ref cur, -arc, arc, cr);
            Bez(ctx, ref cur, -c, d, -(b + c), d, -(a + b + c), d);
            Line(ctx, ref cur, new Point(x0 + p, y0 + h));
            // bottom-left
            Bez(ctx, ref cur, -a, 0, -(a + b), 0, -(a + b + c), -d);
            Arc(ctx, ref cur, -arc, -arc, cr);
            Bez(ctx, ref cur, -d, -c, -d, -(b + c), -d, -(a + b + c));
            Line(ctx, ref cur, new Point(x0, y0 + p));
            // top-left
            Bez(ctx, ref cur, 0, -a, 0, -(a + b), d, -(a + b + c));
            Arc(ctx, ref cur, arc, -arc, cr);
            Bez(ctx, ref cur, c, -d, b + c, -d, a + b + c, -d);
        }
        g.Freeze();
        return g;
    }

    static double Rad(double deg) => deg * Math.PI / 180;

    static void Bez(StreamGeometryContext ctx, ref Point cur, double x1, double y1, double x2, double y2, double x3, double y3)
    {
        var o = cur;
        cur = new Point(o.X + x3, o.Y + y3);
        ctx.BezierTo(new Point(o.X + x1, o.Y + y1), new Point(o.X + x2, o.Y + y2), cur, true, true);
    }

    static void Arc(StreamGeometryContext ctx, ref Point cur, double dx, double dy, double r)
    {
        cur = new Point(cur.X + dx, cur.Y + dy);
        ctx.ArcTo(cur, new Size(r, r), 0, false, SweepDirection.Clockwise, true, true);
    }

    static void Line(StreamGeometryContext ctx, ref Point cur, Point to)
    {
        cur = to;
        ctx.LineTo(to, true, true);
    }
}
