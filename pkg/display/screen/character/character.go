package character

import (
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type character struct{}

func New() display.Screen {
	return &character{}
}

func (s *character) Render(display int, m *display.Manager) image.Image {
	state, _ := m.GetState(display)
	d, ok := state.Data.(CharacterData)
	if !ok {
		return draw.ErrorScreen("UI Character")
	}

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "UI Character")
	draw.Sprite(img, 10, 30, d.Character)

	return img
}

func init() {
	display.Register(display.ScreenUICharacter, New())
}
