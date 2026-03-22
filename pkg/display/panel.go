package display

import "image"

type Panel interface {
	DisplayDraw(channel int, img image.Image) error
}
