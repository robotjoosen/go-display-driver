package menu

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type menu struct {
	data MenuData
}

func New(data MenuData) screens.Screen {
	return &menu{}
}

func (s *menu) Render(data any) image.Image {
	s.data = data.(MenuData)
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "Menu")

	y := 30
	for i, item := range s.data.Items {
		prefix := "  "
		if i == s.data.Selected {
			prefix = "> "
		}
		draw.Text(img, 5, y, 115, 13, fmt.Sprintf("%s%s", prefix, item))
		y += 12
	}

	return img
}
