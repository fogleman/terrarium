package main

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/fogleman/gg"
	"github.com/fogleman/maps"
	"github.com/fogleman/terrarium"
	"github.com/umahmood/haversine"
)

const (
	Z = 13

	// Iceland
	// Lat0, Lng0 = 66.750108, -25.731168
	// Lat1, Lng1 = 62.881496, -12.699955

	// Grand Canyon
	Lat0, Lng0 = 36.477988, -112.726473
	Lat1, Lng1 = 35.940449, -111.561530

	// Mount Everest
	// Lat0, Lng0 = 28.413539, 86.467738
	// Lat1, Lng1 = 27.543224, 87.400420

	URLTemplate    = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
	CacheDirectory = "cache"
	MaxDownloads   = 16
)

func main() {
	min := terrarium.LatLng(Lat0, Lng0)
	max := terrarium.LatLng(Lat1, Lng1)
	bounds := maps.Bounds{maps.Point(min), maps.Point(max)}
	line := maps.NewPolyline([]maps.Point{
		maps.Point{min.X, min.Y},
		maps.Point{max.X, min.Y},
		maps.Point{max.X, max.Y},
		maps.Point{min.X, max.Y},
		maps.Point{min.X, min.Y},
	})
	shape := maps.Shape{bounds, []*maps.Polyline{line}, nil}
	shapes := []maps.Shape{shape}

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

	lo := math.MaxFloat64
	hi := -lo
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			tile, err := cache.GetTile(Z, x, y)
			if err != nil {
				panic(err)
			}
			lo = math.Min(lo, tile.MinElevation)
			hi = math.Max(hi, tile.MaxElevation)
		}
	}
	fmt.Println(lo, hi)

	lo = -1

	fmt.Println("stitching tiles...")
	const TileSize = terrarium.TileSize
	w := (x1 - x0 + 1) * TileSize
	h := (y1 - y0 + 1) * TileSize
	im := image.NewGray16(image.Rect(0, 0, w, h))
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			tile, err := cache.GetTile(Z, x, y)
			if err != nil {
				panic(err)
			}
			tile.MaskShapes(shapes)
			tx0 := (x - x0) * TileSize
			ty0 := (y - y0) * TileSize
			tx1 := tx0 + TileSize
			ty1 := ty0 + TileSize
			r := image.Rect(tx0, ty0, tx1, ty1)
			draw.DrawMask(im, r, tile.AsGray16(lo, hi), image.ZP, tile.Mask, image.ZP, draw.Src)
		}
	}

	px0 := w
	py0 := h
	px1 := 0
	py1 := 0
	for i, v := range im.Pix {
		if v > 0 {
			y := (i / 2) / w
			x := (i / 2) % w
			if x < px0 {
				px0 = x
			}
			if y < py0 {
				py0 = y
			}
			if x > px1 {
				px1 = x
			}
			if y > py1 {
				py1 = y
			}
		}
	}
	trimmed := im.SubImage(image.Rect(px0+1, py0+1, px1-1, py1-1))

	_, lrKm := haversine.Distance(
		haversine.Coord{Lat: Lat0, Lon: Lng0},
		haversine.Coord{Lat: Lat0, Lon: Lng1})
	_, tbKm := haversine.Distance(
		haversine.Coord{Lat: Lat0, Lon: Lng0},
		haversine.Coord{Lat: Lat1, Lon: Lng0})
	xMeters := lrKm * 1000
	yMeters := tbKm * 1000
	zMeters := hi - lo
	xMetersPerPixel := xMeters / float64(px1-px0)
	yMetersPerPixel := yMeters / float64(py1-py0)
	avgMetersPerPixel := (xMetersPerPixel + yMetersPerPixel) / 2
	zScale := zMeters / avgMetersPerPixel
	// fmt.Println(xMeters, yMeters, zMeters)
	// fmt.Println(xMetersPerPixel, yMetersPerPixel)
	fmt.Println(zScale)

	gg.SavePNG("out.png", trimmed)

}
