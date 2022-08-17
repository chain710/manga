package serve

import (
	"errors"
	"strings"
)

type Config struct {
	Addr             string
	Debug            bool
	BaseURI          string
	DSN              string
	ArchiveCacheSize int
	PageCacheSize    int
	ThumbCacheSize   int
	PrefetchImages   int
	PrefetchQueue    int
	ThumbWidth       int
	ThumbHeight      int
}

func (c *Config) Validate() error {
	if c.Addr == "" {
		return errors.New("addr required")
	}

	if c.DSN == "" {
		return errors.New("dsn required")
	}
	if c.ArchiveCacheSize <= 0 {
		return errors.New("invalid archive cache size")
	}
	if c.PageCacheSize <= 0 {
		return errors.New("invalid image cache size")
	}
	if c.ThumbCacheSize <= 0 {
		return errors.New("invalid thumb cache size")
	}
	if c.PrefetchImages > 0 && c.PrefetchQueue <= 0 {
		return errors.New("invalid prefetch queue")
	}
	if c.ThumbWidth < 100 && c.ThumbHeight < 100 {
		return errors.New("thumb too small")
	}
	return nil
}

func (c *Config) GetBaseURI(path string) string {
	if c.BaseURI == "" {
		return path
	}

	return strings.TrimRight(c.BaseURI, "/") + path
}
