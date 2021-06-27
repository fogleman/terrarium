package main

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"

	"github.com/fogleman/gg"
	"github.com/fogleman/terrarium"
)

const (
	Steps     = 100
	Size      = 1600
	Padding   = 0
	LineWidth = 1
)

func main() {
	src, err := gg.LoadImage(os.Args[1])
	if err != nil {
		panic(err)
	}

	gray, _ := ensureGray16(src)
	w := gray.Bounds().Size().X
	h := gray.Bounds().Size().Y
	a := gray16Grid(gray)

	var paths []terrarium.Path
	for i := 0; i < 65535; i += Steps {
		z := float64(i)
		p := terrarium.Slice(a, w, h, z+1e-7)
		if len(p) > 0 {
			fmt.Println("z:", z, len(p))
			paths = append(paths, p...)
		}
	}

	fmt.Println("rendering image...")
	im := renderPaths(paths, Size, Padding, LineWidth)

	fmt.Println("writing png...")
	gg.SavePNG("out.png", im)
}

func ensureGray16(im image.Image) (*image.Gray16, bool) {
	switch im := im.(type) {
	case *image.Gray16:
		return im, true
	default:
		dst := image.NewGray16(im.Bounds())
		draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
		return dst, false
	}
}

func gray16Grid(im *image.Gray16) []float64 {
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	grid := make([]float64, w*h)
	index := 0
	for y := 0; y < h; y++ {
		i := im.PixOffset(0, y)
		for x := 0; x < w; x++ {
			a := int(im.Pix[i]) << 8
			b := int(im.Pix[i+1])
			value := (a | b)
			grid[index] = float64(value)
			index++
			i += 2
		}
	}
	return grid
}

func grayToArray(gray *image.Gray) []float64 {
	w := gray.Bounds().Size().X
	h := gray.Bounds().Size().Y
	a := make([]float64, w*h)
	index := 0
	for y := 0; y < h; y++ {
		i := gray.PixOffset(0, y)
		for x := 0; x < w; x++ {
			a[index] = float64(gray.Pix[i])
			index++
			i++
		}
	}
	return a
}

func renderPaths(paths []terrarium.Path, size, pad int, lw float64) image.Image {
	x0 := paths[0][0].X
	x1 := paths[0][0].X
	y0 := paths[0][0].Y
	y1 := paths[0][0].Y
	for _, path := range paths {
		for _, p := range path {
			if p.X < x0 {
				x0 = p.X
			}
			if p.X > x1 {
				x1 = p.X
			}
			if p.Y < y0 {
				y0 = p.Y
			}
			if p.Y > y1 {
				y1 = p.Y
			}
		}
	}
	pw := x1 - x0
	ph := y1 - y0
	sx := float64(size-pad*2) / pw
	sy := float64(size-pad*2) / ph
	scale := math.Min(sx, sy)
	dc := gg.NewContext(int(pw*scale)+pad*2, int(ph*scale)+pad*2)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.Translate(float64(pad), float64(pad))
	dc.Scale(scale, scale)
	dc.Translate(-x0, -y0)
	for _, path := range paths {
		dc.NewSubPath()
		for _, p := range path {
			dc.LineTo(p.X, p.Y)
		}
	}
	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(lw)
	dc.Stroke()
	return dc.Image()
}
