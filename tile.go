package terrarium

import (
	"image"
	"math"

	"github.com/fogleman/gg"
	"github.com/fogleman/maps"
)

const TileSize = 256

func TileXY(z int, p Point) IntPoint {
	f := TileXYFloat(z, p)
	x := int(math.Floor(f.X))
	y := int(math.Floor(f.Y))
	return IntPoint{x, y}
}

func TileXYFloat(z int, p Point) Point {
	lat := p.Y * math.Pi / 180
	n := math.Pow(2, float64(z))
	x := (p.X + 180) / 360 * n
	y := (1 - math.Log(math.Tan(lat)+(1/math.Cos(lat)))/math.Pi) / 2 * n
	return Point{x, y}
}

func TileLatLng(z, x, y int) Point {
	n := math.Pow(2, float64(z))
	lng := float64(x)/n*360 - 180
	lat := math.Atan(math.Sinh(math.Pi*(1-2*float64(y)/n))) * 180 / math.Pi
	return Point{lng, lat}
}

func imageToElevation(im *image.RGBA) []float64 {
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	buf := make([]float64, w*h)
	index := 0
	for y := 0; y < h; y++ {
		i := im.PixOffset(0, y)
		for x := 0; x < w; x++ {
			r := float64(im.Pix[i+0])
			g := float64(im.Pix[i+1])
			b := float64(im.Pix[i+2])
			meters := (r*256 + g + b/256) - 32768
			buf[index] = meters
			index += 1
			i += 4
		}
	}
	return buf
}

type Tile struct {
	Z, X, Y      int
	W, H         int
	Image        *image.RGBA
	Mask         *image.Alpha
	Elevation    []float64
	MinElevation float64
	MaxElevation float64
}

func newTile(z, x, y int, im image.Image) *Tile {
	rgba, _ := ensureRGBA(im)
	w := rgba.Bounds().Size().X
	h := rgba.Bounds().Size().Y
	elevation := imageToElevation(rgba)
	lo := elevation[0]
	hi := elevation[0]
	for _, e := range elevation {
		if e < lo {
			lo = e
		}
		if e > hi {
			hi = e
		}
	}
	return &Tile{z, x, y, w, h, rgba, nil, elevation, lo, hi}
}

func (tile *Tile) MaskShapes(shapes []maps.Shape) {
	tile.Mask = tile.renderMask(shapes)
}

func (tile *Tile) ContourLines(z float64) []Path {
	return tile.contourLines(z, nil)
}

func (tile *Tile) MaskedContourLines(z float64) []Path {
	return tile.contourLines(z, tile.Mask)
}

func (tile *Tile) contourLines(z float64, mask *image.Alpha) []Path {
	if z < tile.MinElevation || z > tile.MaxElevation {
		return nil
	}
	nw := TileLatLng(tile.Z, tile.X, tile.Y)
	se := TileLatLng(tile.Z, tile.X+1, tile.Y+1)
	pairs := slice(tile.Elevation, tile.W, tile.H, z+1e-7)
	if mask != nil {
		maskedPairs := pairs[:0]
		for _, p := range pairs {
			if mask.AlphaAt(int(p.A.X), int(p.A.Y)).A != 255 {
				continue
			}
			if mask.AlphaAt(int(p.B.X), int(p.B.Y)).A != 255 {
				continue
			}
			maskedPairs = append(maskedPairs, p)
		}
		pairs = maskedPairs
	}
	for i, p := range pairs {
		x0 := nw.X + (se.X-nw.X)*(p.A.X/TileSize)
		y0 := nw.Y + (se.Y-nw.Y)*(p.A.Y/TileSize)
		x1 := nw.X + (se.X-nw.X)*(p.B.X/TileSize)
		y1 := nw.Y + (se.Y-nw.Y)*(p.B.Y/TileSize)
		pairs[i] = pair{Point{x0, y0}, Point{x1, y1}}
	}
	return joinPairs(pairs)
}

func (tile *Tile) MaskedElevation() []float64 {
	elevation := make([]float64, len(tile.Elevation))
	copy(elevation, tile.Elevation)
	mask := tile.Mask
	if mask != nil {
		sentinel := tile.MinElevation - 1
		i := 0
		for y := 0; y < TileSize; y++ {
			j := mask.PixOffset(0, y)
			for x := 0; x < TileSize; x++ {
				if mask.Pix[j] != 255 {
					elevation[i] = sentinel
				}
				i++
				j++
			}
		}
	}
	return elevation
}

func (tile *Tile) renderMask(shapes []maps.Shape) *image.Alpha {
	nw := TileLatLng(tile.Z, tile.X, tile.Y)
	se := TileLatLng(tile.Z, tile.X+1, tile.Y+1)
	dc := gg.NewContext(TileSize, TileSize)
	for _, shape := range shapes {
		for _, line := range shape.Lines {
			dc.NewSubPath()
			for _, p := range line.Points {
				x := (p.X - nw.X) / (se.X - nw.X) * TileSize
				y := (p.Y - nw.Y) / (se.Y - nw.Y) * TileSize
				dc.LineTo(x, y)
			}
		}
	}
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	return dc.AsMask()
}
