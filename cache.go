package terrarium

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const TileSize = 256

type Cache struct {
	URLTemplate  string
	Directory    string
	MaxDownloads int
	sem          chan int
	wg           *sync.WaitGroup
}

func NewCache(urlTemplate, directory string, maxDownloads int) *Cache {
	sem := make(chan int, maxDownloads)
	wg := &sync.WaitGroup{}
	return &Cache{urlTemplate, directory, maxDownloads, sem, wg}
}

func (cache *Cache) EnsureTile(z, x, y int) {
	path := cache.tilePath(z, x, y)
	if _, err := os.Stat(path); err == nil {
		return
	}
	cache.wg.Add(1)
	go cache.tileWorker(z, x, y)
}

func (cache *Cache) Wait() {
	cache.wg.Wait()
}

func (cache *Cache) GetTileImage(z, x, y int) (image.Image, error) {
	path := cache.tilePath(z, x, y)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func (cache *Cache) GetTileElevation(z, x, y int) ([]float64, error) {
	im, err := cache.GetTileImage(z, x, y)
	if err != nil {
		return nil, err
	}
	rgba := im.(*image.RGBA)
	buf := make([]float64, TileSize*TileSize)
	index := 0
	for y := 0; y < TileSize; y++ {
		i := rgba.PixOffset(0, y)
		for x := 0; x < TileSize; x++ {
			r := float64(rgba.Pix[i+0])
			g := float64(rgba.Pix[i+1])
			b := float64(rgba.Pix[i+2])
			meters := (r*256 + g + b/256) - 32768
			buf[index] = meters
			index += 1
			i += 4
		}
	}
	return buf, nil
}

func (cache *Cache) tileURL(z, x, y int) string {
	url := cache.URLTemplate
	url = strings.Replace(url, "{z}", strconv.Itoa(z), -1)
	url = strings.Replace(url, "{x}", strconv.Itoa(x), -1)
	url = strings.Replace(url, "{y}", strconv.Itoa(y), -1)
	return url
}

func (cache *Cache) tilePath(z, x, y int) string {
	path := fmt.Sprintf("%d/%d/%d.png", z, x, y)
	path = filepath.Join(cache.Directory, path)
	return path
}

func (cache *Cache) tileDir(z, x, y int) string {
	path := cache.tilePath(z, x, y)
	dir, _ := filepath.Split(path)
	return dir
}

func (cache *Cache) tileWorker(z, x, y int) {
	defer cache.wg.Done()
	cache.sem <- 1
	err := cache.downloadTile(z, x, y)
	<-cache.sem
	if err != nil {
		panic(err)
	}
}

func (cache *Cache) downloadTile(z, x, y int) error {
	os.MkdirAll(cache.tileDir(z, x, y), os.ModePerm)
	path := cache.tilePath(z, x, y)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	url := cache.tileURL(z, x, y)
	response, err := http.Get(url)
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	defer fmt.Println(path)
	return err
}
