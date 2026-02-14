package draw

import (
	"image"
	"image/color"
	"log/slog"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	white = color.Gray{Y: 255}
)

// only straight lines for now
func Line(img *image.Gray, x1, y1, x2, y2 int) {
	switch {
	case x1 == x2:
		horizontalLine(img, x1, y1, y2, white)
	case y1 == y2:
		verticalLine(img, y1, x1, x2, white)
	default:
		slog.Warn("implement bresenham algorithm here")
	}
}

func Circle(img *image.Gray, x, y, r int) {
	circle(img, x, y, r, 0.0, math.Pi*2, white)
}

func Rectangle(img *image.Gray, x, y, w, h int) {
	rectangle(img, x, y, w, h, white)
}

func RectangleRoundedBorders(img *image.Gray, x, y, w, h, r int) {
	horizontalLine(img, y, x+r, (x+w)-r, white)
	horizontalLine(img, y+h-1, +1+r, (x+w)-r, white)
	verticalLine(img, x, y+r, (y+h)-r, white)
	verticalLine(img, x+w-1, y+r, (y+h)-r, white)

	quater := math.Pi / 2
	circle(img, (x+r)-1, y+r-1, r, quater*2, quater*3, white)
	circle(img, x+(w-r), y+r-1, r, quater*3, quater*4, white)
	circle(img, x+(w-r), y+(h-r), r, 0.0, quater, white)
	circle(img, (x+r)-1, y+(h-r), r, quater, quater*2, white)
}

func Text(img *image.Gray, x, y, w, h int, text string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{white},
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}

	d.DrawString(text)
}

func circle(img *image.Gray, x, y, r int, start, end float64, c color.Gray) {
	start += 0.01 // prevent drawing a rogue pixel
	for o := start; o < end; o += 0.04 {
		px := x + int(float64(r)*math.Cos(o))
		py := y + int(float64(r)*math.Sin(o))
		img.SetGray(px, py, c)
	}
}

func rectangle(img *image.Gray, x, y, w, h int, c color.Gray) {
	rect := image.Rect(x, y, x+w, y+h)
	for x := rect.Min.X; x < rect.Max.X; x++ {
		img.SetGray(x, rect.Min.Y, c)
		img.SetGray(x, rect.Max.Y-1, c)
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		img.SetGray(rect.Min.X, y, c)
		img.SetGray(rect.Max.X-1, y, c)
	}
}

func horizontalLine(img *image.Gray, y, x1, x2 int, c color.Gray) {
	p1 := x1
	p2 := x2
	if p1 > p2 {
		p1 = x2
		p2 = x1
	}

	for p := range p2 - p1 {
		img.SetGray(p1+p, y, c)
	}
}

func verticalLine(img *image.Gray, x, y1, y2 int, c color.Gray) {
	p1 := y1
	p2 := y2
	if p1 > p2 {
		p1 = y2
		p2 = y1
	}

	for p := range p2 - p1 {
		img.SetGray(x, p1+p, c)
	}
}
