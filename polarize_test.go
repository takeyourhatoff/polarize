package main

import (
	"image"
	"image/color"
	"testing"

	"github.com/takeyourhatoff/hsv"
)

func BenchmarkPolarimetricImage_addSample(b *testing.B) {
	r := image.Rect(0, 0, 1000, 1000)
	img := newPolarimetricImage(r)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			img.addSample(0, image.Opaque)
		}
	})
}

func BenchmarkPolarimetricImage_addYCbCrSample(b *testing.B) {
	r := image.Rect(0, 0, 1000, 1000)
	img := newPolarimetricImage(r)
	sample := image.NewYCbCr(r, image.YCbCrSubsampleRatio420)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			img.addSample(0, sample)
		}
	})
}

func BenchmarkPolarimetricImage_At(b *testing.B) {
	r := image.Rect(0, 0, 1000, 1000)
	img := newPolarimetricImage(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				_ = img.At(x, y)
			}
		}
	}
}

func TestPolarimetricImage(t *testing.T) {
	r := image.Rect(0, 0, 3, 1)
	pol := newPolarimetricImage(r)
	sample := image.NewGray(r)
	for x := 0; x < 3; x++ {
		sample.Set(x, 0, color.White)
		pol.addSample(x, sample)
		sample.Set(x, 0, color.Black)
	}
	exp := []color.Color{
		hsv.HSV{H: 0 / 3.0, S: 2 / 3.0, V: 1 / 3.0},
		hsv.HSV{H: 1 / 3.0, S: 2 / 3.0, V: 1 / 3.0},
		hsv.HSV{H: 2 / 3.0, S: 2 / 3.0, V: 1 / 3.0},
	}
	for x, c := range exp {
		if colorEq(pol.At(x, 0), c) == false {
			t.Errorf("x=%v, pol.At(x, 0) = %#v, exp[x] = %#v\n", x, pol.At(x, 0), c)
		}
	}
}

func setYCbCr(i *image.YCbCr, x, y int, c color.Color) {
	Y := color.GrayModel.Convert(c).(color.Gray).Y
	j := i.YOffset(x, y)
	i.Y[j] = Y
}

func TestPolarimetricYCbCrImage(t *testing.T) {
	r := image.Rect(0, 0, 3, 1)
	pol := newPolarimetricImage(r)
	sample := image.NewYCbCr(r, image.YCbCrSubsampleRatio420)
	for x := 0; x < 3; x++ {
		setYCbCr(sample, x, 0, color.White)
		pol.addSample(x, sample)
		setYCbCr(sample, x, 0, color.Black)
	}
	exp := []color.Color{
		hsv.HSV{H: 0 / 3.0, S: 2 / 3.0, V: 1 / 3.0},
		hsv.HSV{H: 1 / 3.0, S: 2 / 3.0, V: 1 / 3.0},
		hsv.HSV{H: 2 / 3.0, S: 2 / 3.0, V: 1 / 3.0},
	}
	for x, c := range exp {
		if colorEq(pol.At(x, 0), c) == false {
			t.Errorf("x=%v, pol.At(x, 0) = %#v, exp[x] = %#v\n", x, pol.At(x, 0), c)
		}
	}
}

func colorEq(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
