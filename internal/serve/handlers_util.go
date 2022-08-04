package serve

import (
	"bytes"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/imagehelper"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/tasks"
	"github.com/chain710/manga/internal/types"
	"github.com/chain710/manga/internal/util"
	"github.com/gin-gonic/gin"
)

func (h *handlers) getVolume(ctx *gin.Context, vid int64) (*db.Volume, error) {
	vol := h.volumesCache.Get(vid)
	if vol != nil {
		log.Debugf("get volume hit cache: %d", vid)
		return vol, nil
	}
	vol, err := h.database.GetVolume(ctx, db.GetVolumeOptions{ID: vid})
	if err != nil {
		log.Errorf("get volume %d error: %s", vid, err)
		return nil, err
	}

	h.volumesCache.Set(vid, vol)
	log.Debugf("get volume hit database: %d", vid)
	return vol, nil
}

type getImageOptions struct {
	prefetch    int
	updateCache bool
}

// getImage get image from cache or disk(archive file). page starts from 1
func (h *handlers) getImage(vol *db.Volume, page int, options getImageOptions) (*types.Image, error) {
	logger := log.With("vol_id", vol.ID, "page", page)
	if page < 1 || page > len(vol.Files) {
		log.Debugf("invalid page: %d", page)
		return nil, errInvalidRequest
	}
	fh, err := util.FileHash(vol.Path)
	if err != nil {
		logger.Errorf("get volume hash error: %s", err)
		return nil, err
	}

	k := cache.ImageKey{Volume: fh, Page: page}
	img, ok := h.imagesCache.Get(k)
	if ok {
		logger.Debugf("get image hit cache")
		return &img, nil
	}

	// cache miss, extract from archive
	archive, err := arc.Open(vol.Path, arc.OpenWithCache(h.archiveCache, fh),
		arc.OpenSkipReadingFiles())
	if err != nil {
		logger.Errorf("open archive %s error: %s", vol.Path, err)
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer archive.Close()
	pageFile := vol.Files[page-1]
	data, err := archive.ReadFileAt(pageFile.Offset)
	if err != nil {
		logger.Errorf("read page file error: %s", err)
		return nil, err
	}

	config, format, err := imagehelper.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		logger.Errorf("decode image config error: %s", err)
		return nil, err
	}

	img = types.Image{
		Data:   data,
		Format: format,
		H:      config.Height,
		W:      config.Width,
	}

	if options.updateCache {
		h.imagesCache.Set(k, img)
	}

	if options.prefetch > 0 {
		h.prefetchImages(vol, fh, page, options.prefetch)
	}
	return &img, nil
}

func (h *handlers) prefetchImages(vol *db.Volume, volHash string, offset int, limit int) {
	if h.imagePrefetch == nil {
		return
	}
	end := offset + limit
	var imgPositions []tasks.ImageLocation
	for i := offset; i < len(vol.Files) && i < end; i++ {
		imgPositions = append(imgPositions,
			// i is index of vol.Files, Page = i+1
			tasks.ImageLocation{Page: i + 1, Offset: vol.Files[i].Offset})
	}

	if !h.imagePrefetch.Add(vol.Path, volHash, imgPositions...) {
		log.Warnf("prefetch %s(%d-%d) not added", vol.Path, offset, end)
	}
}

func (h *handlers) bindAndValidJSON(ctx *gin.Context, req types.Validator) error {
	if err := ctx.ShouldBindJSON(req); err != nil {
		log.Debugf("bind json error: %s", err)
		return errInvalidRequest
	}

	if err := req.Validate(); err != nil {
		log.Debugf("validate request error: %s", err)
		return errInvalidRequest
	}

	return nil
}
