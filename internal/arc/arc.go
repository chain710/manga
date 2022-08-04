package arc

import (
	"github.com/chain710/manga/internal/log"
	"github.com/gen2brain/go-unarr"
	"io"
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

	a, err := unarr.NewArchive(path)
	if err != nil {
		return nil, err
	}

	var files []File
	if !opt.SkipReadingFiles {
		for {
			entryErr := a.Entry()
			if entryErr != nil {
				if entryErr == io.EOF {
					break
				}
				return nil, entryErr
			}

			files = append(files, File{
				name:    a.Name(),
				offset:  a.Offset(),
				modTime: a.ModTime(),
				size:    a.Size(),
			})
		}
	}

	ra := &realArchive{
		path:  path,
		impl:  a,
		files: files,
	}

	if opt.Cache != nil {
		inCache, added := opt.Cache.SetAndGet(opt.CacheKey, ra)
		if !added {
			log.Warnw("duplicate open archive in cache, should close", "key", opt.CacheKey, "path", path)
			_ = ra.Close()
		}
		return inCache, nil
	}
	return ra, nil
}

type Archive interface {
	Path() string
	GetFiles() []File
	ReadFileAt(offset int64) ([]byte, error)
	GetFile(path string) (*File, error)
	Close() error
}
