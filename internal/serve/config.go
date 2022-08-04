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
	VolumeCacheSize  int
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
	if c.VolumeCacheSize <= 0 {
		return errors.New("invalid volume cache size")
	}
	return nil
}

func (c *Config) GetBaseURI(path string) string {
	if c.BaseURI == "" {
		return path
	}

	return strings.TrimRight(c.BaseURI, "/") + path
}
