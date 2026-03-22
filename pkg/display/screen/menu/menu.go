package menu

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type menu struct{}

func New() display.Screen {
	return &menu{}
}

func (s *menu) Render(display int, m *display.Manager) image.Image {
	state, _ := m.GetState(display)
	menuData := MenuData{}
	if md, ok := state.Data.(MenuData); ok {
		menuData = md
	}
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, "Menu")

	y := 30
	for i, item := range menuData.Items {
		prefix := "  "
		if i == menuData.Selected {
			prefix = "> "
		}
		draw.Text(img, 5, y, 115, 13, fmt.Sprintf("%s%s", prefix, item))
		y += 12
	}

	return img
}

func (s *menu) HandleSelect(display int, m *display.Manager) {
	state, ok := m.GetState(display)
	if !ok {
		return
	}

	menuData := MenuData{}
	if md, ok := state.Data.(MenuData); ok {
		menuData = md
	}

	if menuData.Selected < 0 || menuData.Selected >= len(menuData.Items) {
		return
	}

	_ = menuData.Items[menuData.Selected]
}

func init() {
	display.Register(display.ScreenMenu, New())
}
