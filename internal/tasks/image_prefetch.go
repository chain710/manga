package tasks

import (
	"bytes"
	"context"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/imagehelper"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/types"
	"github.com/chain710/manga/internal/util"
)

type ImagePrefetchOption func(*ImagePrefetch)

func NewImagePrefetch(imageCache *cache.Images, archiveCache *arc.ArchiveCache, queueSize int) *ImagePrefetch {
	return &ImagePrefetch{
		ch:           make(chan prefetchImageItem, queueSize),
		imageCache:   imageCache,
		archiveCache: archiveCache,
	}
}

type ImageLocation struct {
	Page   int
	Offset int64
}

type prefetchImageItem struct {
	fileHash string
	path     string
	images   []ImageLocation
}

// ImagePrefetch prefetch images and store in cache
type ImagePrefetch struct {
	ch           chan prefetchImageItem
	imageCache   *cache.Images
	archiveCache *arc.ArchiveCache
}

func (i *ImagePrefetch) Start(ctx context.Context) {
	log.Infof("start image prefetch...")
	go i.workLoop()

	<-ctx.Done()
	log.Debugf("stopping prefetch channel")
	close(i.ch)
}

func (i *ImagePrefetch) Add(path, fileHash string, images ...ImageLocation) bool {
	select {
	case i.ch <- prefetchImageItem{path: path, fileHash: fileHash, images: images}:
		return true
	default:
		return false
	}
}

func (i *ImagePrefetch) workLoop() {
	for pitem := range i.ch {
		i.prefetch(pitem)
	}
}

func (i *ImagePrefetch) prefetch(item prefetchImageItem) {
	a, err := arc.Open(item.path, arc.OpenWithCache(i.archiveCache, item.fileHash))
	if err != nil {
		log.Errorf("open archive %s error: %s", item.path, err)
		return
	}

	//goland:noinspection GoUnhandledErrorResult
	defer a.Close()
	for _, pos := range item.images {
		k := types.PageKey{Volume: item.fileHash, Page: pos.Page}
		if i.imageCache.Has(k) {
			continue
		}
		data, err := a.ReadFileAt(pos.Offset)
		if err != nil {
			continue
		}

		cfg, format, err := imagehelper.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			continue
		}

		img := types.Image{
			Data:   data,
			Hash:   util.ImageHash(data),
			Format: format,
			W:      cfg.Width,
			H:      cfg.Height,
		}
		i.imageCache.Set(k, img)
		log.Debugf("prefetch vol %s/%d, size=%d, wh=(%d, %d)", k.Volume, k.Page, len(data), cfg.Width, cfg.Height)
	}

}
