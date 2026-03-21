package network

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type network struct {
	data NetworkStatusData
}

func New(data NetworkStatusData) screens.Screen {
	return &network{}
}

func (s *network) Render(data any) image.Image {
	s.data = data.(NetworkStatusData)
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	draw.Text(img, 10, 10, 100, 13, s.data.ID)
	draw.Text(img, 10, 30, 100, 13, "Network Status")

	y := 30
	for _, iface := range s.data.Interfaces {
		draw.Text(img, 5, y, 115, 13, fmt.Sprintf("%s %d/%d", iface.Name, iface.Rx/(1024*1024), iface.Tx/(1024*1024)))
		y += 12
	}

	return img
}
