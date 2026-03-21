package character

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type character struct {
	data CharacterData
}

func New(data CharacterData) screens.Screen {
	return &character{}
}

func (s *character) Render(data any) image.Image {
	s.data = data.(CharacterData)
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "UI Character")
	draw.Sprite(img, 10, 30, s.data.Character)

	return img
}
