package main

import (
	"bufio"
	"fmt"
	"image"
	"math"
	"os"
	"runtime"

	"github.com/fogleman/gg"
	"github.com/fogleman/maps"
	"github.com/fogleman/terrarium"
)

const (
	Step          = 250
	Z             = 9
	Size          = 2048
	Padding       = 0
	LineWidth     = 1
	HistogramStep = 100

	Country = ""
	State   = "CO"
	County  = ""
	STATEFP = ""

	Lat, Lng   = 43.482486, -110.780319
	M          = 0.5
	Lat0, Lng0 = Lat - 0.1*M, Lng - 0.175*M
	Lat1, Lng1 = Lat + 0.1*M, Lng + 0.175*M

	CountryShapefile = "ne_10m_admin_0_countries/wgs84.shp"
	// CountryShapefile = "TM_WORLD_BORDERS-0.3/wgs84.shp"
	StateShapefile  = "cb_2016_us_state_500k/wgs84.shp"
	CountyShapefile = "cb_2016_us_county_500k/wgs84.shp"

	URLTemplate    = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
	CacheDirectory = "cache"
	// URLTemplate    = "https://maps.wikimedia.org/osm-intl/{z}/{x}/{y}.png"
	// CacheDirectory = "cache-osm"
	MaxDownloads = 16
)

func loadShapes() ([]maps.Shape, error) {
	var result []maps.Shape
	if Country != "" {
		shapes, err := maps.LoadShapefile(CountryShapefile,
			maps.NewShapeTagFilter("NAME", Country))
		if err != nil {
			return nil, err
		}
		result = append(result, shapes...)
	} else if State != "" {
		shapes, err := maps.LoadShapefile(StateShapefile,
			maps.NewShapeTagFilter("STUSPS", State))
		if err != nil {
			return nil, err
		}
		result = append(result, shapes...)
	} else if County != "" {
		shapes, err := maps.LoadShapefile(CountyShapefile,
			maps.NewShapeTagFilter("NAME", County),
			maps.NewShapeTagFilter("STATEFP", STATEFP))
		if err != nil {
			return nil, err
		}
		result = append(result, shapes...)
	} else {
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
		result = append(result, shape)
	}
	return result, nil
}

func main() {
	shapes, err := loadShapes()
	if err != nil {
		panic(err)
	}
	if len(shapes) == 0 {
		fmt.Println("no shapes!")
		return
	}
	bounds := maps.BoundsForShapes(shapes...)
	min := terrarium.Point(bounds.Min)
	max := terrarium.Point(bounds.Max)
	fmt.Println(min)
	fmt.Println(max)

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

	// {
	// 	const TileSize = terrarium.TileSize
	// 	w := (x1 - x0 + 1) * TileSize
	// 	h := (y1 - y0 + 1) * TileSize
	// 	im := image.NewRGBA(image.Rect(0, 0, w, h))
	// 	for y := y0; y < y1; y++ {
	// 		for x := x0; x < x1; x++ {
	// 			tile, err := cache.GetTile(Z, x, y)
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			tile.MaskShapes(shapes)
	// 			tx0 := (x - x0) * TileSize
	// 			ty0 := (y - y0) * TileSize
	// 			tx1 := x0 + TileSize
	// 			ty1 := y0 + TileSize
	// 			r := image.Rect(tx0, ty0, tx1, ty1)
	// 			draw.DrawMask(im, r, tile.Image, image.ZP, tile.Mask, image.ZP, draw.Src)
	// 		}
	// 	}
	// 	gg.SavePNG("stitched.png", im)
	// 	return
	// }

	fmt.Println("processing tiles...")
	jobs := make(chan job, n)
	results := make(chan result, n)
	wn := runtime.NumCPU()
	for wi := 0; wi < wn; wi++ {
		go worker(cache, jobs, results)
	}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			jobs <- job{Z, x, y, shapes}
		}
	}
	close(jobs)
	var paths []terrarium.Path
	h := make(histogram)
	for i := 0; i < n; i++ {
		r := <-results
		paths = append(paths, r.Paths...)
		h.Update(r.Histogram)
	}
	h.Print()

	fmt.Println("projecting paths...")
	proj := maps.NewMercatorProjection()
	proj.InvertY = true
	for _, path := range paths {
		for i, p := range path {
			q := proj.Project(maps.Point(p))
			path[i].X = q.X
			path[i].Y = q.Y
		}
	}
	for _, shape := range shapes {
		for _, line := range shape.Lines {
			var path terrarium.Path
			for _, p := range line.Points {
				q := proj.Project(maps.Point(p))
				path = append(path, terrarium.Point(q))
			}
			paths = append(paths, path)
		}
	}

	fmt.Println("rendering image...")
	im := renderPaths(paths, Size, Padding, LineWidth)

	fmt.Println("writing png...")
	gg.SavePNG("out.png", im)

	fmt.Println("writing axi...")
	saveAxi("out.axi", paths)
}

func saveAxi(filename string, paths []terrarium.Path) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for _, path := range paths {
		for i, p := range path {
			if i != 0 {
				fmt.Fprintf(w, " ")
			}
			fmt.Fprintf(w, "%g,%g", p.X, p.Y)
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

type job struct {
	Z, X, Y int
	Shapes  []maps.Shape
}

type result struct {
	Paths     []terrarium.Path
	Histogram histogram
}

type histogram map[int]int

func (h histogram) Update(other histogram) {
	for k, v := range other {
		h[k] += v
	}
}

func (h histogram) Print() {
	total := 0
	for z := -10000; z <= 10000; z += HistogramStep {
		count := h[z]
		total += count
		if count > 0 {
			fmt.Println(z, count)
		}
	}
	fmt.Println(total)
}

func worker(cache *terrarium.Cache, in chan job, out chan result) {
	for j := range in {
		tile, err := cache.GetStitchedTile(j.Z, j.X, j.Y)
		if err != nil {
			panic(err)
		}
		tile.MaskShapes(j.Shapes)
		var paths []terrarium.Path
		for z := -10000; z < 10000; z += Step {
			p := tile.MaskedContourLines(float64(z + 1))
			paths = append(paths, p...)
		}
		h := make(histogram)
		for _, e := range tile.MaskedElevation() {
			if e < tile.MinElevation {
				continue
			}
			h[int(e/HistogramStep)*HistogramStep]++
		}
		out <- result{paths, h}
	}
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
	// dc.Identity()
	// dc.DrawCircle(float64(dc.Width()/2), float64(dc.Height()/2), 8)
	// dc.SetRGBA(1, 0, 0, 0.9)
	// dc.Fill()
	return dc.Image()
}
