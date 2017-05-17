package main

import (
	"image"
	"image/color"

	"github.com/takeyourhatoff/hsv"
)

type polarimetricImage struct {
	Pix     []polarimetricPixel
	Stride  int
	Rect    image.Rectangle
	Samples int
}

func newPolarimetricImage(r image.Rectangle, samples int) *polarimetricImage {
	w, h := r.Dx(), r.Dy()
	buf := make([]polarimetricPixel, w*h)
	return &polarimetricImage{buf, w, r, samples}
}

type polarimetricPixel struct {
	maxy    uint8
	avgv, h float32
}

func (p *polarimetricPixel) addSample(m, n int, c color.Color) {
	y := color.GrayModel.Convert(c).(color.Gray).Y
	if y > p.maxy {
		p.maxy = y
		p.h = float32(m) / float32(n)
	}
	p.avgv += float32(y) / (255 * float32(n))
}

func (p *polarimetricPixel) RGBA() (r, g, b, a uint32) {
	maxv := float32(p.maxy) / 255
	return hsv.HSV{
		H: p.h,
		S: abs(maxv - p.avgv),
		V: p.avgv,
	}.RGBA()
}

func (p *polarimetricImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := p.PixOffset(x, y)
	return &p.Pix[i]
}

func (p *polarimetricImage) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x - p.Rect.Min.X)
}

func (p *polarimetricImage) Bounds() image.Rectangle { return p.Rect }

func (p *polarimetricImage) ColorModel() color.Model { return hsv.HSVModel }

func (p *polarimetricImage) addSample(m int, img image.Image) {
	b := img.Bounds()
	b = b.Intersect(p.Bounds())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := p.PixOffset(x, y)
			p.Pix[i].addSample(m, p.Samples, img.At(x, y))
		}
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0 // return correctly abs(-0)
	}
	return x
}
