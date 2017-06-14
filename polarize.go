package polarize

import (
	"image"
	"image/color"
	"sync"

	"github.com/takeyourhatoff/polarize/internal/hsv"
)

type Image struct {
	pix      []polarimetricPixel
	stride   int
	rect     image.Rectangle
	lineMu   []sync.Mutex
	initOnce sync.Once
	samples  int
}

func (p *Image) init(r image.Rectangle) {
	w, h := r.Dx(), r.Dy()
	p.pix = make([]polarimetricPixel, w*h)
	p.stride = w
	p.rect = r
	p.lineMu = make([]sync.Mutex, h)
}

func (p *Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.rect)) {
		return hsv.HSV{}
	}
	i := p.PixOffset(x, y)
	return p.pix[i].finalize(p.samples)
}

func (p *Image) PixOffset(x, y int) int {
	return (y-p.rect.Min.Y)*p.stride + (x - p.rect.Min.X)
}

func (p *Image) Bounds() image.Rectangle { return p.rect }

func (p *Image) ColorModel() color.Model { return hsv.HSVModel }

func (p *Image) AddSample(n int, img image.Image) {
	p.initOnce.Do(func() {
		p.init(img.Bounds())
	})
	b := img.Bounds()
	b = b.Intersect(p.Bounds())
	var f func(int, int, int, int)
	switch iimg := img.(type) {
	case yCbCrImage:
		f = func(i, n, x, y int) { p.pix[i].addYCbCrSample(n, iimg.YCbCrAt(x, y)) }
	default:
		f = func(i, n, x, y int) { p.pix[i].addSample(n, iimg.At(x, y)) }
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		p.lineMu[y-b.Min.Y].Lock()
		for x := b.Min.X; x < b.Max.X; x++ {
			i := p.PixOffset(x, y)
			f(i, n, x, y)
		}
		p.lineMu[y-b.Min.Y].Unlock()
	}
	p.samples++
}

type yCbCrImage interface {
	YCbCrAt(x, y int) color.YCbCr
}

type polarimetricPixel struct {
	maxy uint8
	n    int
	sumy int64
}

func (p *polarimetricPixel) addYCbCrSample(n int, c color.YCbCr) {
	if c.Y > p.maxy {
		p.maxy = c.Y
		p.n = n
	}
	p.sumy += int64(c.Y)
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
