package hsv

import (
	"image"
	"image/color"
	"math"
)

// NoHue is the Hue value used in a HSV color when there is no hue.
const NoHue = -1

// HSV represents a fully opaque HSV color with each component in the range [0..1]
type HSV struct {
	H, S, V float32
}

// RGBA implements color.Color for HSV
func (c HSV) RGBA() (uint32, uint32, uint32, uint32) {
	if c.S == 0 {
		vi := f2i(c.V)
		return vi, vi, vi, 0xffff
	}
	if c.H == 1 {
		c.H = 0
	}
	c.H *= 6
	i := int(c.H)
	f := c.H - float32(i)
	aa := c.V * (1 - c.S)
	bb := c.V * (1 - (c.S * f))
	cc := c.V * (1 - (c.S * (1 - f)))
	switch i {
	case 0:
		return f2i(c.V), f2i(cc), f2i(aa), 0xffff
	case 1:
		return f2i(bb), f2i(c.V), f2i(aa), 0xffff
	case 2:
		return f2i(aa), f2i(c.V), f2i(cc), 0xffff
	case 3:
		return f2i(aa), f2i(bb), f2i(c.V), 0xffff
	case 4:
		return f2i(cc), f2i(aa), f2i(c.V), 0xffff
	case 5:
		return f2i(c.V), f2i(aa), f2i(bb), 0xffff
	default:
		return 0, 0, 0, 0
	}
}

func f2i(f float32) uint32 {
	return uint32(f * math.MaxUint16)
}

func hsvModel(c color.Color) color.Color {
	if _, ok := c.(HSV); ok {
		return c
	}
	ir, ig, ib, _ := c.RGBA()
	r := float32(ir) / math.MaxUint16
	g := float32(ig) / math.MaxUint16
	b := float32(ib) / math.MaxUint16

	max := max(r, g, b)
	min := min(r, g, b)
	d := max - min

	var h, s, v float32
	v = max
	if max != 0 {
		s = d / max
	}

	if s == 0 {
		h = NoHue
	} else {
		switch {
		case r == max:
			h = (g - b) / d
		case g == max:
			h = 2 + (b-r)/d
		case b == max:
			h = 4 + (r-g)/d
		}
		h /= 6
		if h < 0 {
			h += 1
		}
	}
	return HSV{h, s, v}
}

// HSVModel is the color.Model for HSV colors
var HSVModel = color.ModelFunc(hsvModel)

func max(a, b, c float32) float32 {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func min(a, b, c float32) float32 {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

// Saturate returns an image with the saturation ajusted by n.
// 0 means no saturation, ∞ means full saturation.
func Saturate(i image.Image, n float64) image.Image {
	return saturate{i, n}
}

type saturate struct {
	image.Image
	n float64
}

func (s saturate) At(x, y int) color.Color {
	c := HSVModel.Convert(s.Image.At(x, y)).(HSV)
	c.S = float32(sigmoid(s.n * float64(c.S)))
	return c
}

// sigmoid squashes values from [-∞..∞] to [-1..1] with a sigmoid function
func sigmoid(n float64) float64 {
	return 2/(1+math.Exp(-n)) - 1
}
