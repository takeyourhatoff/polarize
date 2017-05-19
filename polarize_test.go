package main

import (
	"image"
	"testing"
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
