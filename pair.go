package terrarium

type pair struct {
	A, B Point
}

func joinPairs(pairs []pair) []Path {
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
		result = append(result, path)
	}
	return result
}