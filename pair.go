package terrarium

type pair struct {
	A, B Point
}

func pairsToPaths(pairs []pair) []Path {
	paths := make([]Path, len(pairs))
	for i, p := range pairs {
		paths[i] = Path{p.A, p.B}
	}
	return paths
}

func joinPairs(pairs []pair) []Path {
	// return pairsToPaths(pairs)
	lookup := make(map[Point][]Point, len(pairs))
	for _, pair := range pairs {
		lookup[pair.A] = append(lookup[pair.A], pair.B)
	}
	var result []Path
	for len(lookup) > 0 {
		var p Point
		for p = range lookup {
			break
		}
		var path Path
		for {
			path = append(path, p)
			if qs, ok := lookup[p]; ok {
				q, a := qs[0], qs[1:]
				if len(a) == 0 {
					delete(lookup, p)
				} else {
					lookup[p] = a
				}
				p = q
			} else {
				break
			}
		}
		// if len(path) < 10 {
		// 	a := path[0]
		// 	b := path[len(path)-1]
		// 	if a.Distance(b) < 1e-6 {
		// 		continue
		// 	}
		// }
		result = append(result, path)
	}
	return result
}
