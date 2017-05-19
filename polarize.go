package main

import (
	"image"
	"image/color"
	"sync"

	"github.com/takeyourhatoff/hsv"
)

type polarimetricImage struct {
	Pix     []polarimetricPixel
	Stride  int
	Rect    image.Rectangle
	lineMu  []sync.Mutex
	samples int
}

func newPolarimetricImage(r image.Rectangle) *polarimetricImage {
	w, h := r.Dx(), r.Dy()
	buf := make([]polarimetricPixel, w*h)
	mu := make([]sync.Mutex, h)
	return &polarimetricImage{buf, w, r, mu, 0}
}

type polarimetricPixel struct {
	maxy    uint8
	sumv, h float32
}

func (p *polarimetricPixel) addSample(n int, c color.Color) {
	y := color.GrayModel.Convert(c).(color.Gray).Y
	if y > p.maxy {
		p.maxy = y
		p.h = float32(n)
	}
	p.sumv += float32(y) / 255
}

func (p *polarimetricPixel) finalize(n int) hsv.HSV {
	maxv := float32(p.maxy) / 255
	avgv := p.sumv / float32(n)
	return hsv.HSV{
		H: p.h / float32(n),
		S: abs(maxv - avgv),
		V: avgv,
	}
}

func (p *polarimetricImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return hsv.HSV{}
	}
	i := p.PixOffset(x, y)
	return p.Pix[i].finalize(p.samples)
}

func (p *polarimetricImage) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x - p.Rect.Min.X)
}

func (p *polarimetricImage) Bounds() image.Rectangle { return p.Rect }

func (p *polarimetricImage) ColorModel() color.Model { return hsv.HSVModel }

func (p *polarimetricImage) addSample(n int, img image.Image) {
	b := img.Bounds()
	b = b.Intersect(p.Bounds())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		p.lineMu[y-b.Min.Y].Lock()
		for x := b.Min.X; x < b.Max.X; x++ {
			i := p.PixOffset(x, y)
			p.Pix[i].addSample(n, img.At(x, y))
		}
		p.lineMu[y-b.Min.Y].Unlock()
	}
	p.samples++
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
