package serve

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/imagehelper"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/tasks"
	"github.com/chain710/manga/internal/types"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

const (
	sortBookLatest     = "latest"
	sortBookRecentRead = "read"

	filterBookMustWithProgress = "with_progress_strict"
	filterBookWithProgress     = "with_progress_relax"
)

var (
	sortBooks = map[string]string{
		sortBookLatest:     "update_at desc",
		sortBookRecentRead: "read_volume_at desc",
	}
	filterBooks = map[string]string{
		filterBookMustWithProgress: db.ListBooksRightJoinProgress,
		filterBookWithProgress:     db.ListBooksLeftJoinProgress,
	}
)

type handlers struct {
	config        Config
	database      db.Interface
	archiveCache  *arc.ArchiveCache
	volumesCache  *cache.Volumes
	imagesCache   *cache.Images
	imagePrefetch *tasks.ImagePrefetch
}

func (h *handlers) registerRoutes(router *gin.Engine) {
	v1 := router.Group(h.config.GetBaseURI("/apis/v1"))
	v1.GET("/library", wrapJSONHandler(h.apiListLibraries))
	v1.POST("/library", wrapJSONHandler(h.apiAddLibrary))
	v1.DELETE("/library/:lib", wrapJSONHandler(h.apiDeleteLibrary))
	v1.PATCH("/library/:lib", wrapJSONHandler(h.apiPatchLibrary))
	v1.GET("/library/:lib", wrapJSONHandler(h.apiGetLibrary))
	v1.GET("/book", wrapJSONHandler(h.apiListBooks))
	v1.PATCH("/book/:bid", wrapJSONHandler(h.apiPatchBook))
	v1.GET("/book/:bid", wrapJSONHandler(h.apiGetBook))
	v1.GET("/book/:bid/thumb", wrapImageHandler(h.apiGetBookThumbnail))
	v1.GET("/volume/:vid", wrapJSONHandler(h.apiGetVolume))
	v1.GET("/volume/:vid/thumb", wrapImageHandler(h.apiGetVolumeThumbnail))
	v1.GET("/volume/:vid/read/:page", wrapImageHandler(h.apiReadPage))
	v1.GET("/volume/:vid/read/:page/thumb", wrapImageHandler(h.apiGetPageThumbnail))
	v1.POST("/batch/volume/progress", wrapJSONHandler(h.apiUpdateVolumeProgress))
}

func (h *handlers) apiListLibraries(ctx *gin.Context) (interface{}, error) {
	return h.database.ListLibraries(ctx)
}

func (h *handlers) apiAddLibrary(ctx *gin.Context) (interface{}, error) {
	var req types.AddLibraryRequest
	if err := h.bindAndValidJSON(ctx, &req); err != nil {
		return nil, err
	}

	lib := db.Library{
		CreateAt: db.NewTime(clk.Now()),
		Name:     req.Name,
		Path:     req.Path,
	}
	if err := h.database.CreateLibrary(ctx, &lib); err != nil {
		log.Errorf("create library %s error: %s", req.Path, err)
		return nil, err
	}

	log.Infof("library %d created: %s", lib.ID, lib.Path)
	return lib, nil
}

func (h *handlers) apiDeleteLibrary(ctx *gin.Context) (interface{}, error) {
	var uriParams struct {
		ID int64 `uri:"lib"`
	}

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, err
	}

	if err := h.database.DeleteLibrary(ctx, db.DeleteLibraryOptions{ID: uriParams.ID}); err != nil {
		log.Errorf("delete library %d error: %s", uriParams.ID, err)
		return nil, err
	}

	log.Infof("library %d deleted", uriParams.ID)
	return nil, nil
}

func (h *handlers) apiPatchLibrary(ctx *gin.Context) (interface{}, error) {
	var uriParams struct {
		ID int64 `uri:"lib"`
	}

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, err
	}

	var patchParams struct {
		Name string `json:"name"`
	}
	if err := ctx.ShouldBindJSON(&patchParams); err != nil {
		log.Debugf("bind json error: %s", err)
		return nil, errInvalidRequest
	}
	opt := db.PatchLibraryOptions{ID: uriParams.ID, Name: patchParams.Name}
	if lib, err := h.database.PatchLibrary(ctx, opt); err != nil {
		log.Errorf("patch lib %d error: %s", uriParams.ID, err)
		return nil, err
	} else {
		return lib, nil
	}
}

func (h *handlers) apiGetLibrary(ctx *gin.Context) (interface{}, error) {
	var uriParams struct {
		ID int64 `uri:"lib"`
	}

	if err := ctx.ShouldBindUri(&uriParams); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, err
	}

	lib, err := h.database.GetLibrary(ctx, uriParams.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errNotFound
		} else {
			log.Errorf("get library %d error: %s", uriParams.ID, err)
		}
		return nil, err
	}

	return lib, nil
}

func (h *handlers) apiListBooks(ctx *gin.Context) (interface{}, error) {
	var queryParam struct {
		LibraryID *int64 `form:"lib"`
		Filter    string `form:"filter"`
		Offset    int    `form:"offset"`
		Limit     int    `form:"limit"`
		Sort      string `form:"sort"`
	}
	if err := ctx.ShouldBindQuery(&queryParam); err != nil {
		log.Debugf("bind query error: %s", err)
		return nil, errInvalidRequest
	}

	if queryParam.Sort == sortBookRecentRead &&
		(queryParam.Filter != filterBookWithProgress && queryParam.Filter != filterBookMustWithProgress) {
		log.Debugf("list recent should filter with progress")
		return nil, errInvalidRequest
	}

	options := db.ListBooksOptions{
		LibraryID: queryParam.LibraryID,
		Offset:    queryParam.Offset,
		Limit:     queryParam.Limit,
	}
	if sort, ok := sortBooks[queryParam.Sort]; ok {
		options.Sort = sort
	}
	if join, ok := filterBooks[queryParam.Filter]; ok {
		options.Join = join
	}

	books, count, err := h.database.ListBooks(ctx, options)
	if err != nil {
		log.Errorf("list books error: %s", err)
		return nil, err
	}

	return map[string]interface{}{
		"books": books,
		"count": count,
	}, nil
}

func (h *handlers) apiPatchBook(ctx *gin.Context) (interface{}, error) {
	var queryParams struct {
		BookID int64 `uri:"bid"`
	}
	var updateRequest struct {
		Name    string `json:"name"`
		Writer  string `json:"writer"`
		Summary string `json:"summary"`
	}

	if err := ctx.ShouldBindUri(&queryParams); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}
	if err := ctx.ShouldBindJSON(&updateRequest); err != nil {
		log.Debugf("bind request error: %s", err)
		return nil, errInvalidRequest
	}

	if book, err := h.database.PatchBook(ctx, db.PatchBookOptions{
		ID:      queryParams.BookID,
		Name:    updateRequest.Name,
		Writer:  updateRequest.Writer,
		Summary: updateRequest.Summary,
	}); err != nil {
		log.Errorf("patch book %d error: %s", queryParams.BookID, err)
		return nil, err
	} else {
		return book, nil
	}
}

func (h *handlers) apiGetVolume(ctx *gin.Context) (interface{}, error) {
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

	prev, next, err := h.database.GetVolumeNeighbour(ctx, db.GetVolumeNeighbourOptions{BookID: vol.BookID, Volume: vol.Volume})
	if err != nil {
		logger.Errorf("get volume neighbour error: %s", err)
		return nil, err
	}

	return types.Volume{Volume: *vol, NextVolumeID: next, PrevVolumeID: prev}, nil
}

func (h *handlers) apiGetBook(ctx *gin.Context) (interface{}, error) {
	var uriParam struct {
		BookID int64 `uri:"bid"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	logger := log.With("book", uriParam.BookID)
	book, err := h.database.GetBook(ctx, db.GetBookOptions{ID: uriParam.BookID})
	if err != nil {
		logger.Errorf("get book error: %s", err)
		return nil, err
	}

	if book == nil {
		logger.Debugf("not found book")
		return nil, errNotFound
	}
	return book, nil
}

func (h *handlers) apiGetBookThumbnail(ctx *gin.Context) (*types.Image, error) {
	var uriParam struct {
		BookID int64 `uri:"bid"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}
	logger := log.With("book", uriParam.BookID)
	thumb, err := h.database.GetBookThumbnail(ctx, uriParam.BookID)
	if err != nil {
		logger.Errorf("get thumb error: %s", err)
		return nil, err
	}

	if thumb == nil {
		logger.Debugf("thumb not found, try default")
		return nil, errNotFound
	}

	cfg, format, err := imagehelper.DecodeConfig(bytes.NewReader(thumb.Thumbnail))
	if err != nil {
		logger.Errorf("decode thumb config error: %s", err)
		return nil, err
	}

	return &types.Image{Data: thumb.Thumbnail, Format: format, H: cfg.Height, W: cfg.Width}, nil
}

func (h *handlers) apiGetVolumeThumbnail(ctx *gin.Context) (*types.Image, error) {
	var uriParam struct {
		VolumeID int64 `uri:"vid"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}
	logger := log.With("volume", uriParam.VolumeID)
	thumb, err := h.database.GetVolumeThumbnail(ctx, db.GetVolumeThumbOptions{ID: uriParam.VolumeID})
	if err != nil {
		logger.Errorf("get thumb error: %s", err)
		return nil, err
	}

	if thumb == nil {
		logger.Debugf("thumb not found, try default")
		return nil, errNotFound
	}

	cfg, format, err := imagehelper.DecodeConfig(bytes.NewReader(thumb.Thumbnail))
	if err != nil {
		logger.Errorf("decode thumb config error: %s", err)
		return nil, err
	}

	return &types.Image{Data: thumb.Thumbnail, Format: format, H: cfg.Height, W: cfg.Width}, nil
}

func (h *handlers) apiReadPage(ctx *gin.Context) (*types.Image, error) {
	var uriParam struct {
		VolumeID int64 `uri:"vid"`
		Page     int   `uri:"page"` // starting from 1
	}

	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	vol, err := h.getVolume(ctx, uriParam.VolumeID)
	if err != nil {
		log.Errorf("get volume %d error: %s", uriParam.VolumeID, err)
		return nil, err
	}

	img, err := h.getImage(vol, uriParam.Page,
		getImageOptions{prefetch: h.config.PrefetchImages, updateCache: true})
	if err != nil {
		return nil, err
	}

	volumeProgress := db.VolumeProgressOptions{
		BookID:   vol.BookID,
		VolumeID: vol.ID,
		Complete: uriParam.Page == len(vol.Files),
		Page:     uriParam.Page,
	}

	// set read progress
	if err := h.database.SetVolumeProgress(ctx, volumeProgress); err != nil {
		log.Errorf("set volume %d progress error: %s", uriParam.VolumeID, err)
	}

	return img, nil
}

func (h *handlers) apiGetPageThumbnail(ctx *gin.Context) (*types.Image, error) {
	var uriParam struct {
		VolumeID int64 `uri:"vid"`
		Page     int   `uri:"page"` // starting from 1
	}

	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	logger := log.With("volume", uriParam.VolumeID, "page", uriParam.Page)
	vol, err := h.getVolume(ctx, uriParam.VolumeID)
	if err != nil {
		logger.Errorf("get volume error: %s", err)
		return nil, err
	}

	img, err := h.getImage(vol, uriParam.Page, getImageOptions{updateCache: true})
	if err != nil {
		return nil, err
	}

	// decode image for thumbnail
	image, err := imaging.Decode(bytes.NewReader(img.Data))
	if err != nil {
		logger.Errorf("decode page image error: %s", err)
		return nil, err
	}
	// TODO cache thumb; w&h come from params
	const width = 140
	const height = 200
	image = imaging.Fit(image, width, height, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, image, imaging.JPEG); err != nil {
		logger.Errorf("encode thumb error: %s", err)
		return nil, err
	}

	return &types.Image{
		Data:   buf.Bytes(),
		Format: "jpeg",
		H:      height,
		W:      width,
	}, nil
}

func (h *handlers) apiUpdateVolumeProgress(ctx *gin.Context) (interface{}, error) {
	var req types.UpdateVolumeProgressRequest
	if err := h.bindAndValidJSON(ctx, &req); err != nil {
		return nil, err
	}

	opt := db.BatchUpdateVolumeProgressOptions{IDs: req.VolumeIDs, Operate: req.Operator}
	if err := h.database.BatchUpdateVolumeProgress(ctx, opt); err != nil {
		log.Errorf("batch update volume progress error: %s", err)
		return nil, err
	}

	return nil, nil
}
