package draw

import "image"

func ErrorScreen(msg string) image.Image {
	img := image.NewGray(image.Rect(0, 0, 128, 64))
	Text(img, 0, 30, 128, 13, msg)
	return img
}
