package main

import (
	"flag"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"

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

	var pimg *polarimetricImage

	for n, name := range flag.Args() {
		log.Printf("processing %q", name)
		img, err := openImage(name)
		if err != nil {
			log.Fatal(err)
		}
		if pimg == nil {
			pimg = newPolarimetricImage(img.Bounds(), len(flag.Args()))
		}
		pimg.addSample(n, img)
	}
	img := hsv.Saturate(pimg, *saturation)
	log.Printf("writing to %q", *out)
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
