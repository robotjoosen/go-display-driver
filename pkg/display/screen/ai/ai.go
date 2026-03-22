package ai

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type ai struct{}

func New() display.Screen {
	return &ai{}
}

func (s *ai) Render(display int, m *display.Manager) image.Image {
	state, _ := m.GetState(display)
	d, ok := state.Data.(AIData)
	if !ok {
		return draw.ErrorScreen("AI Interaction")
	}

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "AI Interaction")
	draw.Text(img, 5, 30, 115, 13, d.Status)
	draw.Text(img, 5, 42, 115, 13, fmt.Sprintf("Query: %s", d.Query))

	return img
}

func init() {
	display.Register(display.ScreenAIInteraction, New())
}
