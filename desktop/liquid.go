package main

import (
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// Liquid Glass, as iOS 27 and macOS 27 draw it: glass on the navigation
// and the controls only, over content that stays opaque; controls whose
// knobs turn to drops of clear glass while they move; adjacent toolbar
// buttons sharing one capsule; prominent actions in tinted glass; and the
// corners of nested shapes concentric.

// Corner radii, concentric from the window in: a shape inset by d inside
// one of radius r has the radius r-d.
const (
	radiusWindow  = 26                           // macOS 27's window corners
	radiusSidebar = radiusWindow - sidebarInset  // the floating sidebar
	radiusItem    = radiusSidebar - sidebarInset // its rows, inset by 8
	radiusCard    = 22                           // content cards and widgets
	radiusRow     = radiusCard - 8               // rows inset in a card
	radiusSection = 18                           // grouped lists
)

// chrome is the glass of bars and controls for this frame, which the
// Appearance settings tune from clear to tinted. It is set by applyTheme.
var chrome glass.Glass

// glassFor is the app's glass at the clarity of the settings: 0 tints it
// with the window's color, as iOS 27's "Tinted", 1 leaves it nearly clear.
func glassFor(p palette, clarity float64) glass.Glass {
	cl := float32(min(max(clarity, 0), 1))
	g := glass.Glass{Clarity: cl}
	if k := 0.45 - cl; k > 0 {
		g.Tint = p.Card.Alpha(k * 1.6)
	}
	return g
}

// pane is chrome glass for a bar or a panel.
func pane() glass.Glass { return chrome }

// control is chrome glass that grows a little while pressed.
func control() glass.Glass {
	g := chrome
	g.Interactive = true
	return g
}

// tinted is the glass of a prominent action: the color shows through it,
// as a tinted glass button.
func tinted(col color) glass.Glass {
	return glass.Glass{Tint: col, Interactive: true}
}

// scrollEdge lays the scroll edge effect over the top (or bottom) h DIPs
// of content scrolling under a floating bar: what passes under it blurs
// progressively and fades into bg, so the bar's glass reads over it.
func scrollEdge(c *ui.Context, h float32, bg color, bottom bool) {
	angle := float32(180)
	e := ui.Box(c).Absolute().Left(0).Right(0).Height(h).PassThrough()
	if bottom {
		angle = 0
		e.Bottom(0)
	} else {
		e.Top(0)
	}
	e.Material(glass.Blur{Radius: 8, Mask: &ui.LinearGradient{From: ui.RGB(0, 0, 0), To: ui.Transparent, Angle: angle}})
	f := ui.Box(c).Absolute().Left(0).Right(0).Height(h).PassThrough().Material(glass.ScrollEdge{Bottom: bottom, Background: bg})
	if bottom {
		f.Bottom(0)
	} else {
		f.Top(0)
	}
}

// lifted is the soft shadow under glass that floats over content, which
// lifts it off an opaque surface, where the glass alone barely shows.
func lifted(e ui.Element) ui.Element {
	return e.Shadow(0, 2, 10, 0, ui.RGBA(0, 0, 0, 0.10))
}

// toolbarGroup puts buttons side by side on one capsule of glass, as
// iOS 27 groups a toolbar's buttons.
func toolbarGroup(c *ui.Context, fn func()) {
	lifted(ui.Row(c).Height(36).Radius(18).AlignItems(ui.Center).Role(ui.RoleToolbar)).Material(pane()).Children(fn)
}

// toolButton is a button of a toolbarGroup.
func toolButton(c *ui.Context, p palette, ic *ui.SVG, label string) ui.Element {
	b := ui.Box(c.Key(label)).Size(40, 36).Radius(18).Center().Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(label).Tooltip(label)
	if b.Pressed() {
		b.Background(p.FillPressed)
	} else if b.Hovered() {
		b.Background(p.Fill.Alpha(0.5))
	}
	b.AnimatePose(ui.Pose{Scale: map[bool]float32{true: 0.92, false: 1}[b.Pressed()]}, ui.Bouncy)
	return b.Children(func() { ui.Icon(c, ic).FontSize(18) })
}

// glassIconButton is a single round button of interactive glass.
func glassIconButton(c *ui.Context, ic *ui.SVG, label string) ui.Element {
	return lifted(ui.Box(c).Size(36, 36).Radius(18).Center().Cursor(ui.CursorPointer)).
		Material(control()).Role(ui.RoleButton).Label(label).Tooltip(label).Children(func() {
		ui.Icon(c, ic).FontSize(18)
	})
}

// prominentButton is a capsule of tinted glass for the main action, with
// white text.
func prominentButton(c *ui.Context, p palette, label string, tint color, disabled bool) ui.Element {
	if disabled {
		tint = tint.Alpha(0.4)
	}
	return ui.Row(c.Key(label)).Height(38).Padding(0, 20).Radius(19).Center().Gap(6).Cursor(ui.CursorPointer).
		Material(tinted(tint)).Role(ui.RoleButton).Label(label).Disabled(disabled).
		TextColor(ui.RGB(255, 255, 255)).FontSize(14).Bold()
}

// roundButton is a circle of tinted glass with a white glyph, as the call
// buttons.
func roundButton(c *ui.Context, ic *ui.SVG, fill color, size float32, label string) ui.Element {
	b := ui.Box(c.Key(label)).Size(size, size).Radius(size / 2).Center().Material(tinted(fill)).
		Cursor(ui.CursorPointer).Role(ui.RoleButton).Label(label).Tooltip(label)
	b.Children(func() {
		ui.Icon(c, ic).FontSize(size * 0.46).TextColor(ui.RGB(255, 255, 255))
	})
	return b
}

// lift is how far a control's knob has lifted into a drop of glass, from
// 0 to 1: while it is pressed, and while it moves to a new place.
func lift(e ui.Element, up bool) float32 {
	return e.Animate("lift", b2f(up), 160*time.Millisecond)
}

func b2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

// knob paints a control's thumb at r: opaque at rest, as iOS's, and a
// drop of clear glass, grown by up to grow DIPs, while lifted. Switches'
// and sliders' knobs are white in the dark too; a segmented control's
// thumb is gray there.
func knob(pt *ui.Painter, p palette, r ui.Rect, up, grow float32, white bool) {
	g := grow * up
	r = ui.Rect{X: r.X - g, Y: r.Y - g*0.7, W: r.W + 2*g, H: r.H + 1.4*g}
	rad := r.H / 2
	solid := ui.RGB(255, 255, 255)
	if p.dark && !white {
		solid = ui.Hex("#5B5B5F")
	}
	if up < 1 {
		pt.Shadow(r, rad, 0, 1+2*(1-up), 4, 0, ui.RGBA(0, 0, 0, 0.16*(1-up)))
	}
	if up > 0 {
		glass.Paint(pt, r, rad, glass.Glass{Style: glass.Clear})
	}
	if up < 1 {
		pt.Fill(r, solid.Alpha(1-up), rad)
	}
}

// segmented is iOS 27's segmented control: a capsule track whose thumb
// slides on a spring to the segment chosen, lifting into glass on the way.
func segmented(c *ui.Context, p palette, sel *int, width float32, labels ...string) ui.Element {
	parts := ui.SegmentedBase(c, sel, len(labels))
	g := parts.Track
	n := float32(len(labels))
	pos := g.AnimateTo("thumb", float32(*sel), ui.Snappy)
	pressed := false
	g.Width(width).Height(32).Padding(2).Radius(16).Background(p.Fill).AlignItems(ui.Stretch)
	g.Children(func() {
		for i, label := range labels {
			s := parts.Segment(i).Grow(1).Basis(0).Center().Cursor(ui.CursorPointer)
			pressed = pressed || s.Pressed()
			on := i == *sel
			s.Children(func() {
				t := ui.Text(c, label).SingleLine().FontSize(13)
				if on {
					t.Bold()
				} else {
					t.TextColor(p.Label.Alpha(0.85))
				}
			})
		}
	})
	moving := pos-float32(*sel) > 0.02 || float32(*sel)-pos > 0.02
	up := lift(g, pressed || moving)
	g.Draw(func(pt *ui.Painter, r ui.Rect) {
		w := (r.W - 4) / n
		knob(pt, p, ui.Rect{X: r.X + 2 + pos*w, Y: r.Y + 2, W: w, H: r.H - 4}, up, 3, false)
	})
	return g
}

// glassSwitch is iOS 27's switch: a green capsule when on, whose
// pill-shaped knob turns to glass and stretches while it is pressed and
// as it slides.
func glassSwitch(c *ui.Context, p palette, on *bool) ui.Element {
	sw := ui.SwitchBase(c, on).Size(44, 26).Radius(13).Cursor(ui.CursorPointer)
	pos := sw.AnimateTo("knob", b2f(*on), ui.Snappy)
	moving := pos > 0.02 && pos < 0.98
	up := lift(sw, sw.Pressed() || moving)
	off := ui.RGBA(120, 120, 128, 0.2)
	if p.dark {
		off = ui.RGBA(120, 120, 128, 0.36)
	}
	sw.Background(off.Mix(p.Green, min(max(pos, 0), 1)))
	sw.Draw(func(pt *ui.Painter, r ui.Rect) {
		kw := float32(26) + 6*up
		kh := r.H - 4
		x := r.X + 2 + pos*(r.W-4-kw)
		knob(pt, p, ui.Rect{X: x, Y: r.Y + 2, W: kw, H: kh}, up, 2, true)
	})
	return sw
}

// glassSlider is iOS 27's slider: a thin track filled with the accent up
// to a capsule knob, which lifts into glass while dragged.
func glassSlider(c *ui.Context, p palette, value *float64, lo, hi float64) ui.Element {
	const kw, kh = 30, 20
	s := ui.SliderBase(c, value, lo, hi).Height(30).PaddingX(kw / 2).Cursor(ui.CursorPointer)
	up := lift(s, s.Pressed())
	accent := c.Theme().Accent
	s.Draw(func(pt *ui.Painter, r ui.Rect) {
		f := float32((*value - lo) / (hi - lo))
		f = min(max(f, 0), 1)
		x0, w := r.X+kw/2, r.W-kw
		y := r.Y + r.H/2
		pt.Fill(ui.Rect{X: x0, Y: y - 3, W: w, H: 6}, p.Fill.Alpha(1.6), 3)
		pt.Fill(ui.Rect{X: x0, Y: y - 3, W: w * f, H: 6}, accent, 3)
		knob(pt, p, ui.Rect{X: x0 + w*f - kw/2, Y: y - kh/2, W: kw, H: kh}, up, 5, true)
	})
	return s
}
