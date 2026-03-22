package network

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type network struct{}

func New() display.Screen {
	return &network{}
}

func (s *network) Render(display int, m *display.Manager) image.Image {
	state, _ := m.GetState(display)
	d, ok := state.Data.(NetworkStatusData)
	if !ok {
		return draw.ErrorScreen("Network Status")
	}

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, d.ID)
	draw.Text(img, 10, 30, 100, 13, "Network Status")

	y := 30
	for _, iface := range d.Interfaces {
		draw.Text(img, 5, y, 115, 13, fmt.Sprintf("%s %d/%d", iface.Name, iface.Rx/(1024*1024), iface.Tx/(1024*1024)))
		y += 12
	}

	return img
}

func init() {
	display.Register(display.ScreenNetworkStatus, New())
}
