package main

import (
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
)

const (
	Width, Height = 16384, 16384
	Iterations = (1<<16 - 1) / Brighten
	Scale = 4.0 / Width
	Brighten = 1024
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

var fractal [Width][Height]uint16

type pixel struct {
	x, y int
	wg   *sync.WaitGroup
}

func compute(x, y int, wg *sync.WaitGroup) {
	fractal[x][y] = mandelbrot(complex(float64(x-Width/2)*Scale, float64(y-Height/2)*Scale))

	wg.Done()
}

var queue = make(chan pixel)

func computeThread() {
	for p := range queue {
		compute(p.x, p.y, p.wg)
	}
}

func main() {
	var wg sync.WaitGroup

	wg.Add(Width * Height)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go computeThread()
	}

	for x := 0; x < Width; x++ {
		for y := 0; y < Height; y++ {
			queue <- pixel{x, y, &wg}
		}
	}
	close(queue)

	wg.Wait()

	img := image.NewGray16(image.Rect(0, 0, Width, Height))
	for y, row := range fractal {
		for x, val := range row {
			img.SetGray16(x, y, color.Gray16{val * Brighten})
		}
	}
	f, _ := os.Create("mandelbrot.png")
	defer f.Close()
	png.Encode(f, img)
}
