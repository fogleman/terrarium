package terrarium

import (
	"image"
	"image/draw"
)

func stitchTile(cache *Cache, z, x, y int) (*image.RGBA, error) {
	n := 1 << uint(z)
	x0 := x
	y0 := y
	x1 := (x0 + 1) % n
	y1 := (y0 + 1) % n

	im00, err := cache.getTileImage(z, x0, y0)
	if err != nil {
		return nil, err
	}

	im01, err := cache.getTileImage(z, x0, y1)
	if err != nil {
		return nil, err
	}

	im10, err := cache.getTileImage(z, x1, y0)
	if err != nil {
		return nil, err
	}

	im11, err := cache.getTileImage(z, x1, y1)
	if err != nil {
		return nil, err
	}

	im := image.NewRGBA(image.Rect(0, 0, TileSize+1, TileSize+1))
	draw.Draw(im, image.Rect(0, 0, TileSize, TileSize), im00, image.ZP, draw.Over)
	draw.Draw(im, image.Rect(0, TileSize, TileSize, TileSize+1), im01, image.ZP, draw.Over)
	draw.Draw(im, image.Rect(TileSize, 0, TileSize+1, TileSize), im10, image.ZP, draw.Over)
	draw.Draw(im, image.Rect(TileSize, TileSize, TileSize+1, TileSize+1), im11, image.ZP, draw.Over)
	return im, nil
}
