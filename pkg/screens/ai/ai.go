package ai

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type ai struct {
	data AIData
}

func New(data AIData) screens.Screen {
	return &ai{}
}

func (s *ai) Render(data any) image.Image {
	s.data = data.(AIData)
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "AI Interaction")
	draw.Text(img, 5, 30, 115, 13, s.data.Status)
	draw.Text(img, 5, 42, 115, 13, fmt.Sprintf("Query: %s", s.data.Query))

	return img
}
