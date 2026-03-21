package loading

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type loading struct {
	data LoadingData
}

func New(data LoadingData) screens.Screen {
	return &loading{}
}

func (s *loading) Render(data any) image.Image {
	s.data = data.(LoadingData)
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, s.data.ID)
	draw.Text(img, 10, 30, 100, 13, fmt.Sprintf("Loading %.0f%%", s.data.Progress))

	progressWidth := int(118 * s.data.Progress / 100)
	draw.RectangleRoundedBorders(img, 5, 45, 123, 58, 3)
	draw.Rectangle(img, 7, 47, 7+progressWidth, 56)

	return img
}
