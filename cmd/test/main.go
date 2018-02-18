package main

import (
	"fmt"

	"github.com/fogleman/terrarium"
)

const (
	URLTemplate    = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
	CacheDirectory = "cache"
	MaxDownloads   = 16
)

func main() {
	// z := 0
	// n := 1 << uint(z)
	// fmt.Println(z, n)

	cache := terrarium.NewCache(URLTemplate, CacheDirectory, MaxDownloads)
	// for y := 0; y < n; y++ {
	// 	for x := 0; x < n; x++ {
	// 		cache.EnsureTile(z, x, y)
	// 	}
	// }
	cache.EnsureTile(14, 3087, 6422)
	cache.Wait()

	fmt.Println(cache.GetTileElevation(14, 3087, 6422))
}
