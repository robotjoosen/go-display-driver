package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	opts := ssd1306.Opts{
		W: 128,
		H: 64,
	}

	dev, err := ssd1306.NewI2C(bus, &opts)
	if err != nil {
		log.Fatal(err)
	}

	img := image.NewGray(image.Rect(0, 0, 128, 64))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(5, 20),
	}
	d.DrawString("Hello, world!")

	dev.Draw(dev.Bounds(), img, image.Point{})
}
