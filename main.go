package main

import (
	"flag"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/takeyourhatoff/hsv"
)

var (
	out        = flag.String("out", "out.jpg", "output location (png/jpeg)")
	saturation = flag.Float64("saturation", 10, "saturation coefficent")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write mem profile to file")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	imgs := make([]image.Image, len(flag.Args()))
	var wg sync.WaitGroup
	for i, name := range flag.Args() {
		wg.Add(1)
		go func(i int, name string) {
			img, err := openImage(name)
			if err != nil {
				log.Fatal(err)
			}
			imgs[i] = img
			wg.Done()
		}(i, name)
	}
	wg.Wait()
	img := Polarimetric(imgs)
	img = hsv.Saturate(img, *saturation)
	img = pCopyToRGBA(img, runtime.NumCPU())
	err := saveImage(*out, img)
	if err != nil {
		log.Fatal(err)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}

func openImage(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func saveImage(name string, i image.Image) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	switch path.Ext(name) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(f, i, nil)
	case ".png":
		err = png.Encode(f, i)
	}
	if err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func pCopyToRGBA(img image.Image, n int) *image.RGBA {
	out := image.NewRGBA(img.Bounds())
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go work(out, img, i, n, &wg)
	}
	wg.Wait()
	return out
}

func work(out draw.Image, in image.Image, i, n int, wg *sync.WaitGroup) {
	b := in.Bounds()
	for y := b.Min.Y + i; y < b.Max.Y; y += n {
		for x := b.Min.X; x < b.Max.X; x++ {
			out.Set(x, y, in.At(x, y))
		}
	}
	wg.Done()
}
