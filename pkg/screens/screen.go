package screens

import "image"

type Screen interface {
	Render(data any) image.Image
}
