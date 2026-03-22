package error

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type error struct{}

func New() display.Screen {
	return &error{}
}

func (s *error) Render(display int, m *display.Manager) image.Image {
	state, _ := m.GetState(display)
	d, ok := state.Data.(ErrorData)
	if !ok {
		return draw.ErrorScreen("Error")
	}

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "Error")
	draw.Text(img, 5, 30, 115, 13, d.ID)
	draw.Text(img, 5, 42, 115, 13, d.Message)

	return img
}

func init() {
	display.Register(display.ScreenError, New())
}
