package ui

import (
	"fmt"
	"image"

	"github.com/robotjoosen/go-display-driver/pkg/draw"
)

func Generate(
	id string,
	state bool,
	cpu float64,
	memory uint64,
	networkRx uint64,
	networkTx uint64,
) image.Image {
	img := image.NewGray(image.Rect(0, 0, 128, 64))

	drawOnlineState(img, state)
	draw.Text(img, 20, 10, 100, 13, id)

	draw.RectangleRoundedBorders(img, 1, 15, 126, 48, 6)
	draw.Text(img, 5, 30, 115, 13, fmt.Sprintf("cpu %.2f", cpu))
	draw.Text(img, 5, 42, 115, 13, fmt.Sprintf("memory %d", memory/(1024*1024)))
	draw.Text(img, 5, 54, 115, 13, fmt.Sprintf("network %d/%d", networkRx/(1024*1024), networkTx/(1024*1024)))

	return img
}

func drawOnlineState(img *image.Gray, online bool) {
	draw.Circle(img, 6, 5, 6)

	if online {
		draw.Circle(img, 6, 5, 2)
		draw.Circle(img, 6, 5, 1)
	}
}
