package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	Size       = 256
	Iterations = (1<<16 - 1) / Brighten
	Brighten   = 1024
)

func mandelbrot(c complex128) uint16 {
	var z complex128

	for i := 0; i < Iterations; i++ {
		z = z*z + c
		if cmplx.IsNaN(z) {
			return uint16(i)
		}
	}

	return Iterations
}

type pixel struct {
	out          *image.Gray16
	x, y         int
	tileX, tileY int64
	tileZoom     uint8
	wg           *sync.WaitGroup
}

var queue = make(chan pixel)

func computeThread() {
	for p := range queue {
		val := mandelbrot(
			complex(
				(float64(p.x)/Size+float64(p.tileX))/float64(uint(1<<p.tileZoom)),
				(float64(p.y)/Size+float64(p.tileY))/float64(uint(1<<p.tileZoom)),
			),
		)
		p.out.SetGray16(p.x, p.y, color.Gray16{val * Brighten})

		p.wg.Done()
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go computeThread()
	}

	log.Fatal(http.ListenAndServe(":6161", nil))
}

func renderTile(w http.ResponseWriter, r *http.Request) {
	components := strings.Split(r.URL.Path, "/")[1:]

	if len(components) != 4 || components[0] != "mandelbrot" || components[3][len(components[3])-4:] != ".png" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	components[3] = components[3][:len(components[3])-4]

	tileX, err := strconv.ParseInt(components[2], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	tileY, err := strconv.ParseInt(components[3], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	tileZoom, err := strconv.ParseUint(components[1], 10, 8)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var wg sync.WaitGroup

	wg.Add(Size * Size)

	img := image.NewGray16(image.Rect(0, 0, Size, Size))

	for x := 0; x < Size; x++ {
		for y := 0; y < Size; y++ {
			queue <- pixel{img, x, y, tileX, tileY, uint8(tileZoom), &wg}
		}
	}

	wg.Wait()

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

func init() {
	http.HandleFunc("/mandelbrot/", renderTile)
}
