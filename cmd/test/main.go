package main

import (
	"image"
	"math"

	"github.com/fogleman/gg"
	"github.com/fogleman/maps"
	"github.com/fogleman/terrarium"
)

const (
	URLTemplate    = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
	CacheDirectory = "cache"
	MaxDownloads   = 16
)

func main() {
	// z := 0
	// n := 1 << uint(z)
	// fmt.Println(z, n)

	cache := terrarium.NewCache(URLTemplate, CacheDirectory, MaxDownloads)
	// for y := 0; y < n; y++ {
	// 	for x := 0; x < n; x++ {
	// 		cache.EnsureTile(z, x, y)
	// 	}
	// }
	cache.EnsureTile(14, 3087, 6422)
	cache.Wait()

	tile, err := cache.GetTile(14, 3087, 6422)
	if err != nil {
		panic(err)
	}

	var paths []terrarium.Path
	for z := 0; z < 5000; z += 25 {
		p := tile.ContourLines(float64(z))
		paths = append(paths, p...)
	}

	proj := maps.NewMercatorProjection()
	for _, path := range paths {
		for i, p := range path {
			q := proj.Project(maps.Point{p.X, p.Y})
			path[i].X = q.X
			path[i].Y = q.Y
		}
	}

	im := renderPaths(paths, 1024, 0, 1)
	gg.SavePNG("out.png", im)
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
