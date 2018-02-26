package terrarium

import "math"

type IntPoint struct {
	X, Y int
}

type Point struct {
	X, Y float64
}

func (a Point) Distance(b Point) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func LatLng(lat, lng float64) Point {
	return Point{lng, lat}
}

type Path []Point

type Bounds struct {
	Min, Max Point
}
