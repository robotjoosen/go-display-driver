package device

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
	"github.com/robotjoosen/go-display-driver/pkg/screens"
)

type device struct {
	data DeviceStatusData
}

func New(data DeviceStatusData) screens.Screen {
	return &device{}
}

func (s *device) Render(data any) image.Image {
	s.data = data.(DeviceStatusData)

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	drawOnlineState(img, s.data.Online)
	draw.Text(img, 20, 10, 100, 13, s.data.ID)

	draw.RectangleRoundedBorders(img, 1, 15, 126, 48, 6)
	draw.Text(img, 5, 30, 115, 13, fmt.Sprintf("cpu %.2f", s.data.CPU))
	draw.Text(img, 5, 42, 115, 13, fmt.Sprintf("memory %d", s.data.Memory/(1024*1024)))
	draw.Text(img, 5, 54, 115, 13, fmt.Sprintf("network %d/%d", s.data.NetworkRx/(1024*1024), s.data.NetworkTx/(1024*1024)))

	return img
}

func drawOnlineState(img *image.Gray, online bool) {
	draw.Circle(img, 6, 5, 6)

	if online {
		draw.Circle(img, 6, 5, 2)
		draw.Circle(img, 6, 5, 1)
	}
}
