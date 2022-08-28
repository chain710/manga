package arc

import (
	"github.com/chain710/manga/internal/log"
	"path/filepath"
	"strings"
)

type OpenOptions struct {
	Cache            *ArchiveCache
	CacheKey         string
	SkipReadingFiles bool
}

type OpenOption func(*OpenOptions)

func OpenWithCache(cache *ArchiveCache, key string) OpenOption {
	return func(options *OpenOptions) {
		if cache != nil {
			options.Cache = cache
			options.CacheKey = key
		}
	}
}

func OpenSkipReadingFiles() OpenOption {
	return func(options *OpenOptions) {
		options.SkipReadingFiles = true
	}
}

func Open(path string, option ...OpenOption) (Archive, error) {
	var opt OpenOptions
	for _, apply := range option {
		apply(&opt)
	}

	if opt.Cache != nil {
		archive := opt.Cache.Get(opt.CacheKey)
		if archive != nil {
			log.Debugw("archive hit cache", "path", path, "key", opt.CacheKey)
			return archive, nil
		}
	}

	archive, err := open(path, opt)
	if err != nil {
		log.Errorf("open archive failed", "path", path, "error", err)
		return nil, err
	}

	if opt.Cache != nil {
		inCache, added := opt.Cache.SetAndGet(opt.CacheKey, archive)
		if !added {
			log.Warnw("duplicate open archive in cache, should close", "key", opt.CacheKey, "path", path)
			_ = archive.Close()
		}
		return inCache, nil
	}
	return archive, nil
}

func open(p string, opt OpenOptions) (Archive, error) {
	ext := strings.ToLower(filepath.Ext(p))
	var tryFunctions []func() (Archive, error)
	switch ext {
	case ".rar":
		tryFunctions = append(tryFunctions, func() (Archive, error) { return newRARArchive(p, opt.SkipReadingFiles) })
		tryFunctions = append(tryFunctions, func() (Archive, error) { return newRealArchive(p, opt.SkipReadingFiles) })
	default:
		tryFunctions = append(tryFunctions, func() (Archive, error) { return newRealArchive(p, opt.SkipReadingFiles) })
		tryFunctions = append(tryFunctions, func() (Archive, error) { return newRARArchive(p, opt.SkipReadingFiles) })
	}

	var ret Archive
	var err error
	for _, f := range tryFunctions {
		ret, err = f()
		if err == nil {
			return ret, nil
		}
	}

	return nil, err
}

type Archive interface {
	Path() string
	GetFiles() []File
	ReadFileAt(offset int64) ([]byte, error)
	GetFile(path string) (*File, error)
	Close() error
}
