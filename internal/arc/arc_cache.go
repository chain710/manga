package arc

import (
	"errors"
	"fmt"
	"github.com/chain710/manga/internal/log"
	lru "github.com/hashicorp/golang-lru"
	"go.uber.org/atomic"
)

func NewArchiveCache(size int) *ArchiveCache {
	ac := new(ArchiveCache)
	cache, err := lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		ac.onEvict(key, value)
	})

	if err != nil {
		panic(err)
	}

	ac.cache = cache
	return ac
}

type ArchiveCache struct {
	cache *lru.Cache
}

func (a *ArchiveCache) Has(key string) bool {
	return a.cache.Contains(key)
}

func (a *ArchiveCache) Get(key string) Archive {
	value, ok := a.cache.Get(key)
	if !ok {
		return nil
	}

	v := value.(*proxyArchive)
	return v.Clone()
}

// Set can not replace existing key, return add ok
func (a *ArchiveCache) Set(key string, archive Archive) bool {
	this := &proxyArchive{Archive: archive, ref: atomic.NewInt32(1)}
	ok, _ := a.cache.ContainsOrAdd(key, this)
	return !ok
}

// SetAndGet same as Set, return <in cache archive, add ok>
func (a *ArchiveCache) SetAndGet(key string, archive Archive) (Archive, bool) {
	this := &proxyArchive{Archive: archive, ref: atomic.NewInt32(1)}
	prev, ok, _ := a.cache.PeekOrAdd(key, this)
	if ok {
		return prev.(*proxyArchive).Clone(), false
	} else {
		return this.Clone(), true
	}
}

func (a *ArchiveCache) Remove(key string) {
	a.cache.Remove(key)
}

func (a *ArchiveCache) onEvict(key, value interface{}) {
	sk := key.(string)
	ca := value.(*proxyArchive)
	log.Debugf("archive %s: %s evicted", sk, ca.Path())
	_ = ca.Close()
}

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type proxyArchive struct {
	noCopy
	Archive
	closed bool
	ref    *atomic.Int32
}

func (c *proxyArchive) Clone() *proxyArchive {
	count := c.ref.Load()
	if count == 0 {
		panic("should not clone 0 ref object")
	}
	c.ref.Add(1)
	return &proxyArchive{
		Archive: c.Archive,
		closed:  false,
		ref:     c.ref,
	}
}

// Close do not really close archive
func (c *proxyArchive) Close() error {
	if c.closed {
		return errors.New("already closed")
	}
	if count := c.ref.Sub(1); count == 0 {
		log.Debugf("close archive %s", c.Archive.Path())
		c.closed = true
		return c.Archive.Close()
	} else if count > 0 {
		c.closed = true
		return nil
	} else {
		panic(fmt.Errorf("close count(%d) < 0; %s", count, c.Archive.Path()))
	}
}
