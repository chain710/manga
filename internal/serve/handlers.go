package serve

import (
	"bytes"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/gin-gonic/gin"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type handlers struct {
	config       Config
	database     db.Interface
	archiveCache *arc.ArchiveCache
	volumesCache *cache.Volumes
}

func (h *handlers) registerRoutes(router *gin.Engine) {
	v1 := router.Group(h.config.GetBaseURI("/apis/v1"))
	v1.GET("/library", wrapJSONHandler(h.listLibraries))
	v1.GET("/volume/:vid/pages", wrapJSONHandler(h.listPages))
	v1.GET("/volume/:vid/read/:page", wrapImageHandler(h.readPage))
}

func (h *handlers) listLibraries(ctx *gin.Context) (interface{}, error) {
	return h.database.ListLibraries(ctx)
}

func (h *handlers) listPages(ctx *gin.Context) (interface{}, error) {
	var uriParam struct {
		VolumeID int64 `uri:"vid"`
	}

	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	logger := log.With("volume", uriParam.VolumeID)
	vol, err := h.getVolume(ctx, uriParam.VolumeID)
	if err != nil {
		logger.Errorf("get volume error: %s", err)
		return nil, err
	}

	return vol.Files, nil
}

func (h *handlers) readPage(ctx *gin.Context) (*imageData, error) {
	var uriParam struct {
		VolumeID int64 `uri:"vid"`
		Page     int   `uri:"page"` // starting from 0
	}

	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	vol, archive, err := h.getVolumeArchive(ctx, uriParam.VolumeID)
	if err != nil {
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer archive.Close()
	if uriParam.Page < 0 || uriParam.Page >= len(vol.Files) {
		log.Debugf("invalid page: %d", uriParam.Page)
		return nil, errInvalidRequest
	}

	logger := log.With("volume", uriParam.VolumeID, "page", uriParam.Page)
	pageFile := vol.Files[uriParam.Page]
	data, err := archive.ReadFileAt(pageFile.Offset)
	if err != nil {
		logger.Errorf("read page file error: %s", err)
		return nil, err
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		logger.Errorf("decode image config error: %s", err)
		return nil, err
	}

	volumeProgress := db.VolumeProgressOptions{
		BookID:   vol.BookID,
		VolumeID: vol.ID,
		Complete: uriParam.Page == len(vol.Files)-1,
		Page:     uriParam.Page,
	}

	// set read progress
	if err := h.database.SetVolumeProgress(ctx, volumeProgress); err != nil {
		logger.Errorf("set volume progress error: %s", err)
	}
	// TODO prefetch images in the background
	return &imageData{data: data, format: format}, nil
}
