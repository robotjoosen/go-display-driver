package startup

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/sprite"
)

type startup struct{}

func New() display.Screen {
	return &startup{}
}

func (s *startup) Render(display int, m *display.Manager) image.Image {
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Sprite(img, 0, 0, sprite.SpriteLogo)

	return img
}

func init() {
	display.Register(display.ScreenStartup, New())
}
