package terrarium

import (
	"image"
	"image/draw"
)

func ensureRGBA(im image.Image) (*image.RGBA, bool) {
	switch im := im.(type) {
	case *image.RGBA:
		return im, true
	default:
		dst := image.NewRGBA(im.Bounds())
		draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
		return dst, false
	}
}
