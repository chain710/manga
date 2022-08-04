package cache

import (
	"github.com/chain710/manga/internal/db"
	lru "github.com/hashicorp/golang-lru"
)

func NewVolumes(size int) *Volumes {
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}
	return &Volumes{cache: cache}
}

type Volumes struct {
	cache *lru.Cache
}

func (a *Volumes) Has(vid int64) bool {
	return a.cache.Contains(vid)
}

func (a *Volumes) Get(vid int64) *db.Volume {
	value, ok := a.cache.Get(vid)
	if !ok {
		return nil
	}
	return value.(*db.Volume).DeepCopy()
}

func (a *Volumes) Set(vid int64, volume *db.Volume) {
	a.cache.Add(vid, volume.DeepCopy())
}

func (a *Volumes) Remove(vid int64) {
	a.cache.Remove(vid)
}
