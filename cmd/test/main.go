package main

import (
	"fmt"
	"image"
	"math"

	"github.com/fogleman/gg"
	"github.com/fogleman/maps"
	"github.com/fogleman/terrarium"
)

const (
	Step    = 200
	Z       = 9
	State   = "UT"
	Country = ""
)

const (
	StateShapefile   = "cb_2016_us_state_500k/wgs84.shp"
	CountryShapefile = "ne_10m_admin_0_countries/wgs84.shp"
	URLTemplate      = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
	CacheDirectory   = "cache"
	MaxDownloads     = 16
)

func loadShapes() []maps.Shape {
	var result []maps.Shape
	if Country != "" {
		shapes, err := maps.LoadShapefile(CountryShapefile)
		if err != nil {
			panic(err)
		}
		filteredShapes := shapes[:0]
		for _, shape := range shapes {
			if shape.Tags["NAME"] == Country {
				filteredShapes = append(filteredShapes, shape)
			}
		}
		result = append(result, filteredShapes...)
	}
	if State != "" {
		shapes, err := maps.LoadShapefile(StateShapefile)
		if err != nil {
			panic(err)
		}
		filteredShapes := shapes[:0]
		for _, shape := range shapes {
			if shape.Tags["STUSPS"] == State {
				filteredShapes = append(filteredShapes, shape)
			}
		}
		result = append(result, filteredShapes...)
	}
	return result
}

func boundsForShapes(shapes []maps.Shape) (min, max terrarium.Point) {
	var x0, y0, x1, y1 float64
	x0 = 180
	x1 = -180
	y0 = 90
	y1 = -90
	for _, shape := range shapes {
		for _, line := range shape.Lines {
			for _, p := range line.Points {
				if p.X < x0 {
					x0 = p.X
				}
				if p.Y < y0 {
					y0 = p.Y
				}
				if p.X > x1 {
					x1 = p.X
				}
				if p.Y > y1 {
					y1 = p.Y
				}
			}
		}
	}
	min = terrarium.Point{x0, y0}
	max = terrarium.Point{x1, y1}
	return
}

func main() {
	shapes := loadShapes()
	min, max := boundsForShapes(shapes)

	p0 := terrarium.TileXY(Z, min)
	p1 := terrarium.TileXY(Z, max)
	x0 := p0.X
	y0 := p0.Y
	x1 := p1.X
	y1 := p1.Y
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	x1++
	y1++
	n := (y1 - y0) * (x1 - x0)
	fmt.Printf("%d tiles\n", n)

	fmt.Println("downloading tiles...")
	cache := terrarium.NewCache(URLTemplate, CacheDirectory, MaxDownloads)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			cache.EnsureTile(Z, x, y)
		}
	}
	cache.Wait()

	fmt.Println("extracting contour lines...")
	var paths []terrarium.Path
	i := 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			tile, err := cache.GetStitchedTile(Z, x, y)
			if err != nil {
				panic(err)
			}
			for z := -500; z < 5000; z += Step {
				p := tile.ContourLines(float64(z))
				paths = append(paths, p...)
			}
			i++
			fmt.Println(i, n)
		}
	}

	fmt.Println("projecting paths...")
	proj := maps.NewMercatorProjection()
	for _, path := range paths {
		for i, p := range path {
			q := proj.Project(maps.Point{p.X, p.Y})
			path[i].X = q.X
			path[i].Y = q.Y
		}
	}
	for _, shape := range shapes {
		for _, line := range shape.Lines {
			for i, p := range line.Points {
				q := proj.Project(maps.Point{p.X, p.Y})
				line.Points[i].X = q.X
				line.Points[i].Y = q.Y
			}
		}
	}

	fmt.Println("rendering output...")
	im := renderPaths(shapes, paths, 4096, 0, 1)
	gg.SavePNG("out.png", im)
}

func renderPaths(shapes []maps.Shape, paths []terrarium.Path, size, pad int, lw float64) image.Image {
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
	dc.InvertY()
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.Translate(float64(pad), float64(pad))
	dc.Scale(scale, scale)
	dc.Translate(-x0, -y0)
	for _, shape := range shapes {
		for _, line := range shape.Lines {
			dc.NewSubPath()
			for _, p := range line.Points {
				dc.LineTo(p.X, p.Y)
			}
		}
	}
	dc.Clip()
	for _, path := range paths {
		dc.NewSubPath()
		for _, p := range path {
			dc.LineTo(p.X, p.Y)
		}
	}
	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(lw)
	dc.Stroke()
	for _, shape := range shapes {
		for _, line := range shape.Lines {
			dc.NewSubPath()
			for _, p := range line.Points {
				dc.LineTo(p.X, p.Y)
			}
		}
	}
	dc.SetLineWidth(lw * 3)
	dc.Stroke()
	return dc.Image()
}
