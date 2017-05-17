package main

import (
	"image"
	"image/color"

	"github.com/takeyourhatoff/hsv"
)

// Polarimetric returns a polarimetric image calculated from the images in photos.
// photos should be a slice of photographs taken with a linear polarization filter.
// Each successive photo in photos should be taken with the polarizing filter rotated by a constant angle.
func Polarimetric(photos []image.Image) image.Image {
	return pol{photos}
}

type pol struct {
	is []image.Image
}

func (p pol) At(x, y int) color.Color {
	len := float32(len(p.is))
	var maxv, avgv, h float32
	for n, i := range p.is {
		c := hsv.HSVModel.Convert(i.At(x, y)).(hsv.HSV)
		if c.V > maxv {
			maxv = c.V
			h = float32(n) / len
		}
		avgv += c.V * (1 / len)
	}
	return hsv.HSV{
		H: h,                // The angle of the most strongly polarized light
		S: abs(maxv - avgv), // The magnitude of the polarization
		V: avgv,             // The average luminance of all input images
	}
}

func (p pol) Bounds() image.Rectangle {
	if len(p.is) == 0 {
		return image.ZR
	}
	bounds := p.is[0].Bounds()
	for _, i := range p.is {
		bounds.Intersect(i.Bounds())
	}
	return bounds
}

func (p pol) ColorModel() color.Model {
	return hsv.HSVModel
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
