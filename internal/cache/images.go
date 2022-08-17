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

func (a *Images) Has(k interface{}) bool {
	return a.cache.Contains(k)
}

func (a *Images) Get(k interface{}) (types.Image, bool) {
	value, ok := a.cache.Get(k)
	if !ok {
		return types.Image{}, false
	}
	return value.(types.Image), true
}

func (a *Images) Set(k interface{}, i types.Image) {
	a.cache.Add(k, i)
}

func (a *Images) Remove(k interface{}) {
	a.cache.Remove(k)
}
