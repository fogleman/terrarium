package terrarium

import (
	"math"
)

type vector struct {
	X, Y, Z float64
}

func (a vector) Dot(b vector) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a vector) Cross(b vector) vector {
	x := a.Y*b.Z - a.Z*b.Y
	y := a.Z*b.X - a.X*b.Z
	z := a.X*b.Y - a.Y*b.X
	return vector{x, y, z}
}

func (a vector) Normalize() vector {
	r := 1 / math.Sqrt(a.X*a.X+a.Y*a.Y+a.Z*a.Z)
	return vector{a.X * r, a.Y * r, a.Z * r}
}

func (a vector) Add(b vector) vector {
	return vector{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a vector) Sub(b vector) vector {
	return vector{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (a vector) MulScalar(b float64) vector {
	return vector{a.X * b, a.Y * b, a.Z * b}
}
