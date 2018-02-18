package terrarium

type IntPoint struct {
	X, Y int
}

type Point struct {
	X, Y float64
}

type Path []Point

func LatLng(lat, lng float64) Point {
	return Point{lng, lat}
}
