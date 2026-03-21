package error

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type error struct {
	data ErrorData
}

func New(data ErrorData) screens.Screen {
	return &error{}
}

func (s *error) Render(data any) image.Image {
	s.data = data.(ErrorData)
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "Error")
	draw.Text(img, 5, 30, 115, 13, s.data.ID)
	draw.Text(img, 5, 42, 115, 13, s.data.Message)

	return img
}
