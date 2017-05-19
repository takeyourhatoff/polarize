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
	maxy uint8
	n    int
	sumy int64
}

func (p *polarimetricPixel) addSample(n int, c color.Color) {
	y := color.GrayModel.Convert(c).(color.Gray).Y
	if y > p.maxy {
		p.maxy = y
		p.n = n
	}
	p.sumy += int64(y)
}

func (p *polarimetricPixel) finalize(m int) hsv.HSV {
	maxv := float32(p.maxy) / 255
	avgv := float32(p.sumy) / (float32(m) * 255)
	return hsv.HSV{
		H: float32(p.n) / float32(m),
		S: maxv - avgv,
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
