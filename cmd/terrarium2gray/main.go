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
	Z = 15

	// Antelope Island
	// Lat, Lng = 40.078190, -111.826498
	// W, H     = 8, 16

	// Yosemite Valley
	Lat, Lng = 37.733614, -119.586598
	W, H     = 15, 10
	// Lat0, Lng0 = 37.685097, -119.674018
	// Lat1, Lng1 = 37.816173, -119.443466

	// Iceland
	// Lat0, Lng0 = 66.750108, -25.731168
	// Lat1, Lng1 = 62.881496, -12.699955
	// Lat0, Lng0 = 66.750108, -25.831168
	// Lat1, Lng1 = 62.881496, -12.599955

	// Grand Canyon
	// Lat0, Lng0 = 36.477988, -112.726473
	// Lat1, Lng1 = 35.940449, -111.561530
	// Lat0, Lng0 = 36.551589, -112.728958
	// Lat1, Lng1 = 35.886162, -111.688370

	// Mount Everest
	// Lat0, Lng0 = 28.413539, 86.467738
	// Lat1, Lng1 = 27.543224, 87.400420

	URLTemplate    = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
	CacheDirectory = "cache"
	MaxDownloads   = 16
)

func boundingBox(lat, lng, widthKm, heightKm float64) (float64, float64, float64, float64) {
	const eps = 1e-3
	_, lrKm := haversine.Distance(
		haversine.Coord{Lat: lat, Lon: lng - eps},
		haversine.Coord{Lat: lat, Lon: lng + eps})
	kmPerLng := lrKm / (2 * eps)
	_, tbKm := haversine.Distance(
		haversine.Coord{Lat: lat - eps, Lon: lng},
		haversine.Coord{Lat: lat + eps, Lon: lng})
	kmPerLat := tbKm / (2 * eps)
	dLng := widthKm / 2 / kmPerLng
	dLat := heightKm / 2 / kmPerLat
	lat0, lat1 := lat-dLat, lat+dLat
	lng0, lng1 := lng-dLng, lng+dLng
	return lat0, lng0, lat1, lng1
}

func main() {
	lat0, lng0, lat1, lng1 := boundingBox(Lat, Lng, W, H)
	min := terrarium.LatLng(lat0, lng0)
	max := terrarium.LatLng(lat1, lng1)
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

	// var result []maps.Shape
	// shapes, err := maps.LoadShapefile("ne_10m_admin_0_countries/wgs84.shp",
	// 	maps.NewShapeTagFilter("NAME", "Iceland"))
	// if err != nil {
	// 	panic(err)
	// }

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
	// fmt.Println(lo, hi)

	// lo = 0
	// hi = 2000

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
		haversine.Coord{Lat: lat0, Lon: lng0},
		haversine.Coord{Lat: lat0, Lon: lng1})
	_, tbKm := haversine.Distance(
		haversine.Coord{Lat: lat0, Lon: lng0},
		haversine.Coord{Lat: lat1, Lon: lng0})
	xMeters := lrKm * 1000
	yMeters := tbKm * 1000
	zMeters := hi - lo
	xMetersPerPixel := xMeters / float64(px1-px0)
	yMetersPerPixel := yMeters / float64(py1-py0)
	avgMetersPerPixel := (xMetersPerPixel + yMetersPerPixel) / 2
	zScale := zMeters / avgMetersPerPixel
	// fmt.Println(xMeters, yMeters, zMeters)
	// fmt.Println(xMetersPerPixel, yMetersPerPixel)
	fmt.Println("width =", trimmed.Bounds().Size().X)
	fmt.Println("height =", trimmed.Bounds().Size().Y)
	fmt.Println("min elevation =", lo)
	fmt.Println("max elevation =", hi)
	fmt.Println("elevation change =", hi-lo)
	fmt.Println("z scale =", zScale)

	gg.SavePNG("out.png", trimmed)

}
