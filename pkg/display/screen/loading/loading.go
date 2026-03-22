package loading

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type loading struct{}

func New() display.Screen {
	return &loading{}
}

func (s *loading) Render(display int, m *display.Manager) image.Image {
	state, _ := m.GetState(display)
	d, ok := state.Data.(LoadingData)
	if !ok {
		return draw.ErrorScreen("Loading")
	}

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, d.ID)
	draw.Text(img, 10, 30, 100, 13, fmt.Sprintf("Loading %.0f%%", d.Progress))

	progressWidth := int(118 * d.Progress / 100)
	draw.RectangleRoundedBorders(img, 5, 45, 123, 58, 3)
	draw.Rectangle(img, 7, 47, 7+progressWidth, 56)

	return img
}

func init() {
	display.Register(display.ScreenLoading, New())
}
