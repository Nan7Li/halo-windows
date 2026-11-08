package main

import (
	"image/png"
	"os"
	"testing"

	"github.com/egoist/mygo/ui"
)

// iconView is the app icon: the Halo ring on a blue to indigo squircle,
// with the margins of macOS's icon grid (824 of 1024).
func iconView(c *ui.Context) {
	c.Root().Background(ui.Transparent)
	ui.Box(c).Fill().Center().Children(func() {
		ui.Box(c).Size(824, 824).Radius(185).Center().
			LinearGradient(ui.LinearGradient{Angle: 160, Oklab: true, Stops: []ui.GradientStop{
				{Color: ui.Hex("#3D8BFF"), At: 0}, {Color: ui.Hex("#5856D6"), At: 1},
			}}).
			Shadow(0, 12, 28, 0, ui.RGBA(0, 0, 0, 0.25)).
			InsetShadow(0, 2, 0, 0, ui.RGBA(255, 255, 255, 0.35)).
			Children(func() {
				ui.Box(c).Size(470, 470).Radius(235).Border(58, ui.RGB(255, 255, 255)).Center().Children(func() {
					ui.Box(c).Size(150, 150).Radius(75).Background(ui.RGB(255, 255, 255))
				})
			})
	})
}

// TestWriteIcon writes resources/icon.png when HALO_WRITE_ICON is set:
//
//	HALO_WRITE_ICON=1 go test -run TestWriteIcon
func TestWriteIcon(t *testing.T) {
	img := ui.Render(iconView, 1024, 1024, 1)
	if img.Bounds().Dx() != 1024 {
		t.Fatalf("icon is %v", img.Bounds())
	}
	if os.Getenv("HALO_WRITE_ICON") == "" {
		return
	}
	if err := os.MkdirAll("resources", 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create("resources/icon.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
