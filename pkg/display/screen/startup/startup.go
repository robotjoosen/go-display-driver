package startup

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type startup struct{}

func New() display.Screen {
	return &startup{}
}

func (s *startup) Render(display int, m *display.Manager) image.Image {
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 20, 20, 100, 13, "Display Driver")
	draw.Text(img, 30, 40, 80, 13, "Initializing...")

	return img
}

func init() {
	display.Register(display.ScreenStartup, New())
}
