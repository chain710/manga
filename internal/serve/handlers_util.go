package serve

import (
	"bytes"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/imagehelper"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/tasks"
	"github.com/chain710/manga/internal/types"
	"github.com/chain710/manga/internal/util"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"image"
)

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

	k := types.PageKey{Volume: fh, Page: page}
	img, ok := h.volumePageCache.Get(k)
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
		Hash:   util.ImageHash(data),
		Format: format,
		H:      config.Height,
		W:      config.Width,
	}

	if options.updateCache {
		h.volumePageCache.Set(k, img)
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

func (h *handlers) cropAsFitThumb(vol *db.Volume, page int, rect image.Rectangle) (*types.Image, error) {
	logger := log.With("volume", vol.ID, "page", page)
	pageImage, err := h.getImage(vol, page, getImageOptions{prefetch: 0, updateCache: false})
	if err != nil {
		logger.Errorf("get image error: %s", err)
		return nil, err
	}

	img, err := imagehelper.DecodeFromBytes(pageImage.Data)
	if err != nil {
		logger.Errorf("decode image error: %s", err)
		return nil, err
	}

	x0 := int(float64(rect.Min.X) / 1000 * float64(pageImage.W))
	y0 := int(float64(rect.Min.Y) / 1000 * float64(pageImage.H))
	x1 := int(float64(rect.Max.X) / 1000 * float64(pageImage.W))
	y1 := int(float64(rect.Max.Y) / 1000 * float64(pageImage.H))
	logger.Debugf("ready to crop image (%d, %d),(%d, %d)", x0, y0, x1, y1)
	cropRect := image.Rect(x0, y0, x1, y1)
	cropImage := imaging.Crop(img, cropRect)
	thumb := imaging.Fit(cropImage, h.config.ThumbWidth, h.config.ThumbHeight, imaging.Lanczos)
	var out bytes.Buffer
	if err := imaging.Encode(&out, thumb, imaging.JPEG); err != nil {
		logger.Errorf("encode crop image error: %s", err)
		return nil, err
	}

	size := cropImage.Rect
	return &types.Image{
		Data:   out.Bytes(),
		Hash:   util.ImageHash(out.Bytes()),
		Format: "jpeg", // TODO use const
		H:      size.Dy(),
		W:      size.Dx(),
	}, nil
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
