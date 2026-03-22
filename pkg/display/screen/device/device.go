package device

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/device"
	"github.com/robotjoosen/go-display-driver/pkg/display"
	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

type deviceScreen struct{}

func New() display.Screen {
	return &deviceScreen{}
}

func (s *deviceScreen) Render(display int, m *display.Manager) image.Image {
	devices := device.GetByType(device.DeviceTypeSBC)

	state, _ := m.GetState(display)
	if len(devices) == 0 {
		return draw.ErrorScreen("No Devices")
	}

	index := state.ListIndex % len(devices)
	d := devices[index]

	img := image.NewGray(image.Rect(0, 0, 128, 64))

	drawOnlineState(img, d.IsOnline())
	draw.Text(img, 20, 10, 100, 13, d.ID())

	draw.RectangleRoundedBorders(img, 1, 15, 126, 48, 6)
	draw.Text(img, 5, 30, 115, 13, fmt.Sprintf("cpu %.2f", d.CPU()))
	draw.Text(img, 5, 42, 115, 13, fmt.Sprintf("memory %d", d.Memory()/(1024*1024)))
	draw.Text(img, 5, 54, 115, 13, fmt.Sprintf("network %d/%d", d.NetworkRx()/(1024*1024), d.NetworkTx()/(1024*1024)))

	m.SetListLength(display, len(devices))

	return img
}

func drawOnlineState(img *image.Gray, online bool) {
	draw.Circle(img, 6, 5, 6)

	if online {
		draw.Circle(img, 6, 5, 2)
		draw.Circle(img, 6, 5, 1)
	}
}

func init() {
	display.Register(display.ScreenDeviceStatus, New())
}
