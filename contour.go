package terrarium

import "github.com/fogleman/fauxgl"

func intersectSegment(z float64, v0, v1 fauxgl.Vector) (fauxgl.Vector, bool) {
	if v0.Z == v1.Z {
		return fauxgl.Vector{}, false
	}
	t := (z - v0.Z) / (v1.Z - v0.Z)
	if t < 0 || t > 1 {
		return fauxgl.Vector{}, false
	}
	v := v0.Add(v1.Sub(v0).MulScalar(t))
	return v, true
}

func intersectTriangle(z float64, pt1, pt2, pt3 fauxgl.Vector) (fauxgl.Vector, fauxgl.Vector, bool) {
	v1, ok1 := intersectSegment(z, pt1, pt2)
	v2, ok2 := intersectSegment(z, pt2, pt3)
	v3, ok3 := intersectSegment(z, pt3, pt1)
	var p1, p2 fauxgl.Vector
	if ok1 && ok2 {
		p1, p2 = v1, v2
	} else if ok1 && ok3 {
		p1, p2 = v1, v3
	} else if ok2 && ok3 {
		p1, p2 = v2, v3
	} else {
		return fauxgl.Vector{}, fauxgl.Vector{}, false
	}
	n := fauxgl.Vector{p1.Y - p2.Y, p2.X - p1.X, 0}
	e1 := pt2.Sub(pt1)
	e2 := pt3.Sub(pt1)
	tn := e1.Cross(e2).Normalize()
	if n.Dot(tn) < 0 {
		return p1, p2, true
	} else {
		return p2, p1, true
	}
}

func slice(grid []float64, w, h int, z float64) []pair {
	var result []pair
	for y := 0; y < h-1; y++ {
		fy := float64(y)
		i0 := y * w
		i1 := i0 + w
		for x := 0; x < w-1; x++ {
			fx := float64(x)
			z0 := grid[i0+x]
			z1 := grid[i0+x+1]
			z2 := grid[i1+x]
			z3 := grid[i1+x+1]
			if z0 < z && z1 < z && z2 < z && z3 < z {
				continue
			}
			if z0 > z && z1 > z && z2 > z && z3 > z {
				continue
			}
			z4 := (z0 + z1 + z2 + z3) / 4
			v0 := fauxgl.Vector{fx, fy, z0}
			v1 := fauxgl.Vector{fx + 1, fy, z1}
			v2 := fauxgl.Vector{fx, fy + 1, z2}
			v3 := fauxgl.Vector{fx + 1, fy + 1, z3}
			v4 := fauxgl.Vector{fx + 0.5, fy + 0.5, z4}
			if p1, p2, ok := intersectTriangle(z, v0, v2, v4); ok {
				pair := pair{Point{p1.X, p1.Y}, Point{p2.X, p2.Y}}
				result = append(result, pair)
			}
			if p1, p2, ok := intersectTriangle(z, v0, v4, v1); ok {
				pair := pair{Point{p1.X, p1.Y}, Point{p2.X, p2.Y}}
				result = append(result, pair)
			}
			if p1, p2, ok := intersectTriangle(z, v1, v4, v3); ok {
				pair := pair{Point{p1.X, p1.Y}, Point{p2.X, p2.Y}}
				result = append(result, pair)
			}
			if p1, p2, ok := intersectTriangle(z, v2, v3, v4); ok {
				pair := pair{Point{p1.X, p1.Y}, Point{p2.X, p2.Y}}
				result = append(result, pair)
			}
		}
	}
	return result
}

func Slice(grid []float64, w, h int, z float64) []Path {
	pairs := slice(grid, w, h, z)
	return joinPairs(pairs)
}
