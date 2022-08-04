package cache

import (
	"github.com/chain710/manga/internal/types"
	lru "github.com/hashicorp/golang-lru"
)

func NewImages(size int) *Images {
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}
	return &Images{cache: cache}
}

// Images stores recent/future visit images
type Images struct {
	cache *lru.Cache
}

type ImageKey struct {
	Volume string // volume's hash
	Page   int
}

func (a *Images) Has(k ImageKey) bool {
	return a.cache.Contains(k)
}

func (a *Images) Get(k ImageKey) (types.Image, bool) {
	value, ok := a.cache.Get(k)
	if !ok {
		return types.Image{}, false
	}
	return value.(types.Image), true
}

func (a *Images) Set(k ImageKey, i types.Image) {
	a.cache.Add(k, i)
}

func (a *Images) Remove(k ImageKey) {
	a.cache.Remove(k)
}
