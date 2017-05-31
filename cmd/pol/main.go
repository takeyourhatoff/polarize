package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/takeyourhatoff/polarize"
	"github.com/takeyourhatoff/polarize/internal/hsv"
)

var (
	out        = flag.String("out", "out.jpg", "output location (png/jpeg)")
	saturation = flag.Float64("saturation", 1, "saturation coefficent")
	numcpu     = flag.Int("numcpu", runtime.NumCPU(), "number of CPU's to utilize, memory usage is proportional to this flag")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write mem profile to file")
)

type sample struct {
	index int
	name  string
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [options] photo.jpg...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	samples := make(chan sample)
	var pimg *polarize.Image
	var wg sync.WaitGroup
	var initPolarimetricImageOnce sync.Once
	for i := 0; i < *numcpu; i++ {
		wg.Add(1)
		go func() {
			for sample := range samples {
				log.Printf("processing %q", sample.name)
				img, err := openImage(sample.name)
				if err != nil {
					log.Fatal(err)
				}
				initPolarimetricImageOnce.Do(func() {
					pimg = polarize.New(img.Bounds())
				})
				pimg.AddSample(sample.index, img)
			}
			wg.Done()
		}()
	}
	for n, name := range flag.Args() {
		samples <- sample{n, name}
	}
	close(samples)
	wg.Wait()

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
