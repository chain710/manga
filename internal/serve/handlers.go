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
	"github.com/chain710/manga/internal/util"
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
	config          Config
	database        db.Interface
	archiveCache    *arc.ArchiveCache
	volumesCache    *cache.Volumes
	volumePageCache *cache.Images
	thumbCache      *cache.Images
	imagePrefetch   *tasks.ImagePrefetch
	libWatcher      *tasks.LibraryWatcher
	thumbScanner    *tasks.ThumbnailScanner
}

func (h *handlers) registerRoutes(router *gin.Engine) {
	v1 := router.Group(h.config.GetBaseURI("/apis/v1"))
	v1.GET("/library", wrapJSONHandler(h.apiListLibraries))
	v1.POST("/library", wrapJSONHandler(h.apiAddLibrary))
	v1.DELETE("/library/:lib", wrapJSONHandler(h.apiDeleteLibrary))
	v1.PATCH("/library/:lib", wrapJSONHandler(h.apiPatchLibrary))
	v1.GET("/library/:lib", wrapJSONHandler(h.apiGetLibrary))
	v1.GET("/library/:lib/scan", wrapJSONHandler(h.apiScanLibrary))
	v1.GET("/book", wrapJSONHandler(h.apiListBooks))
	v1.PATCH("/book/:bid", wrapJSONHandler(h.apiPatchBook))
	v1.GET("/book/:bid", wrapJSONHandler(h.apiGetBook))
	v1.GET("/book/:bid/thumb", wrapImageHandler(h.apiGetBookThumbnail, h.thumbCache))
	v1.POST("/book/:bid/thumb", wrapJSONHandler(h.apiSetBookThumbnail))
	v1.GET("/book/:bid/scan", wrapJSONHandler(h.apiScanBook))
	v1.GET("/volume", wrapJSONHandler(h.apiListVolume))
	v1.GET("/volume/:vid", wrapJSONHandler(h.apiGetVolume))
	// NOTE get/set thumb url path must be exact same to make sure evict thumb cache work
	v1.GET("/volume/:vid/thumb", wrapImageHandler(h.apiGetVolumeThumbnail, h.thumbCache))
	v1.POST("/volume/:vid/thumb", wrapJSONHandler(h.apiSetVolumeThumbnail))
	v1.GET("/volume/:vid/crop/:page/:rect", wrapImageHandler(h.apiCropPage, nil))
	v1.GET("/volume/:vid/read/:page", wrapImageHandler(h.apiReadPage, nil))
	v1.GET("/volume/:vid/read/:page/thumb", wrapImageHandler(h.apiGetPageThumbnail, h.thumbCache))
	v1.POST("/batch/volume/progress", wrapJSONHandler(h.apiUpdateVolumeProgress))
	v1.GET("/fs/listdir", wrapJSONHandler(h.apiListDirectory))
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
	_ = h.libWatcher.AddLibrary(lib)
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

func (h *handlers) apiScanLibrary(ctx *gin.Context) (interface{}, error) {
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

	if nil != h.libWatcher {
		log.Infof("add scan book %d task", uriParams.ID)
		_ = h.libWatcher.AddLibrary(*lib)
	}

	if nil != h.thumbScanner {
		// list volume without thumb
		vols, err := h.database.ListVolumes(ctx, db.ListVolumesOptions{LibraryID: &uriParams.ID, Join: db.VolumeMustNotHaveThumb})
		if err != nil {
			log.Errorf("get volumes of lib %d error: %s", uriParams.ID, err)
		} else {
			h.thumbScanner.ScanVolumes(vols...)
		}
		books, _, err := h.database.ListBooks(ctx, db.ListBooksOptions{LibraryID: &uriParams.ID, Join: db.ListBookWithoutThumbnail})
		if err != nil {
			log.Errorf("list books of lib %d error: %s", uriParams.ID, err)
		} else {
			h.thumbScanner.ScanBook(books...)
		}
	}
	return nil, nil
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

const (
	volumeFilterReading = "reading"
)

func (h *handlers) apiListVolume(ctx *gin.Context) (interface{}, error) {
	queryParams := struct {
		Filter string `form:"filter"`
		Limit  int    `form:"limit"`
	}{}

	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		log.Debugf("bind query error: %s", err)
		return nil, errInvalidRequest
	}

	if queryParams.Filter != volumeFilterReading {
		log.Debugf("invalid request")
		return nil, errInvalidRequest
	}

	opt := db.ListVolumesOptions{Limit: queryParams.Limit}
	switch queryParams.Filter {
	case volumeFilterReading:
		opt.Join = db.VolumeReading
	}
	volumes, err := h.database.ListVolumes(ctx, opt)
	if err != nil {
		log.Errorf("list volumes error: %s", err)
		return nil, err
	}
	return volumes, nil
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
	vol, err := h.database.GetVolume(ctx, db.GetVolumeOptions{ID: uriParam.VolumeID})
	if err != nil {
		logger.Errorf("get volume error: %s", err)
		return nil, err
	}

	opt := db.GetVolumeNeighbourOptions{BookID: vol.BookID, Volume: vol.Volume, VolumeID: uriParam.VolumeID}
	prev, next, err := h.database.GetVolumeNeighbour(ctx, opt)
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
	book, err := h.database.GetBook(ctx, db.GetBookOptions{ID: uriParam.BookID, Join: db.GetBookJoinVolumeProgress})
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

func (h *handlers) apiScanBook(ctx *gin.Context) (interface{}, error) {
	var uriParam struct {
		ID int64 `uri:"bid"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	logger := log.With("book_id", uriParam.ID)
	book, err := h.database.GetBook(ctx, db.GetBookOptions{ID: uriParam.ID})
	if err != nil {
		logger.Errorf("get book error: %s", err)
		return nil, err
	}
	if book == nil {
		logger.Debugf("book not found")
		return nil, errNotFound
	}

	if nil != h.libWatcher {
		logger.Infof("add scan book %d task", uriParam.ID)
		_ = h.libWatcher.AddBook(*book)
	}

	if nil != h.thumbScanner {
		// list volume without thumb
		vols, err := h.database.ListVolumes(ctx, db.ListVolumesOptions{BookID: &uriParam.ID, Join: db.VolumeMustNotHaveThumb})
		if err != nil {
			logger.Errorf("get volumes of book error: %s", err)
		} else {
			h.thumbScanner.ScanVolumes(vols...)
		}

		thumb, err := h.database.GetBookThumbnail(ctx, uriParam.ID)
		if err != nil {
			logger.Errorf("get book thumb error: %s", err)
		} else if thumb == nil {
			h.thumbScanner.ScanBook(*book)
		}
	}

	return nil, nil
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
		logger.Debugf("thumb not found")
		return nil, errNotFound
	}

	cfg, format, err := imagehelper.DecodeConfig(bytes.NewReader(thumb.Thumbnail))
	if err != nil {
		logger.Errorf("decode thumb config error: %s", err)
		return nil, err
	}

	return &types.Image{Data: thumb.Thumbnail, Hash: thumb.Hash, Format: format, H: cfg.Height, W: cfg.Width}, nil
}

func (h *handlers) apiSetBookThumbnail(ctx *gin.Context) (interface{}, error) {
	var uriParam struct {
		BookID int64 `uri:"bid"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	data, err := ctx.GetRawData()
	if err != nil {
		log.Errorf("get raw data error: %s", err)
		return nil, err
	}

	// make sure post data is image using DecodeConfig
	_, format, err := imagehelper.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		log.Errorf("post data is not image: %s", err)
		return nil, err
	}

	if err := h.database.SetBookThumbnail(ctx, db.BookThumbnail{
		ID:        uriParam.BookID,
		Hash:      util.ImageHash(data),
		Thumbnail: data,
	}); err != nil {
		log.Errorf("set book %d thumb error: %s", uriParam.BookID, err)
		return nil, err
	}

	// evict cache
	key := ctx.Request.URL.Path
	h.thumbCache.Remove(key)
	log.Infof("set book %d thumb (format %s) ok. remove cache: %s", uriParam.BookID, format, key)
	return nil, nil
}

func (h *handlers) apiCropPage(ctx *gin.Context) (*types.Image, error) {
	var uriParam struct {
		VolumeID int64  `uri:"vid"`
		Page     int    `uri:"page"`
		Rect     string `uri:"rect"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	rect, err := parseRect(uriParam.Rect)
	if err != nil {
		log.Debugf("parse rect `%s` error: %s", uriParam.Rect, err)
		return nil, errInvalidRequest
	}

	vol, err := h.database.GetVolume(ctx, db.GetVolumeOptions{ID: uriParam.VolumeID})
	if err != nil {
		log.Errorf("get volume %d error: %s", uriParam.VolumeID, err)
		return nil, err
	}

	img, err := h.cropAsFitThumb(vol, uriParam.Page, rect)
	if err != nil {
		log.Errorf("crop as fit thumb error: %s", err)
		return nil, err
	}
	return img, nil
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

	return &types.Image{Data: thumb.Thumbnail, Hash: thumb.Hash, Format: format, H: cfg.Height, W: cfg.Width}, nil
}

func (h *handlers) apiSetVolumeThumbnail(ctx *gin.Context) (interface{}, error) {
	var uriParam struct {
		VolumeID int64 `uri:"vid"`
	}
	if err := ctx.ShouldBindUri(&uriParam); err != nil {
		log.Debugf("bind uri error: %s", err)
		return nil, errInvalidRequest
	}

	data, err := ctx.GetRawData()
	if err != nil {
		log.Errorf("get raw data error: %s", err)
		return nil, err
	}

	// make sure post data is image using DecodeConfig
	_, format, err := imagehelper.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		log.Errorf("post data is not image: %s", err)
		return nil, err
	}

	if err := h.database.SetVolumeThumbnail(ctx, db.VolumeThumbnail{
		ID:        uriParam.VolumeID,
		Hash:      util.ImageHash(data),
		Thumbnail: data,
	}); err != nil {
		log.Errorf("set book %d thumb error: %s", uriParam.VolumeID, err)
		return nil, err
	}

	// evict cache
	key := ctx.Request.URL.Path
	h.thumbCache.Remove(key)
	log.Infof("set volume %d thumb (format %s) ok. remove cache: %s", uriParam.VolumeID, format, key)
	return nil, nil
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

	vol, err := h.database.GetVolume(ctx, db.GetVolumeOptions{ID: uriParam.VolumeID})
	if err != nil {
		log.Errorf("get volume %d error: %s", uriParam.VolumeID, err)
		return nil, err
	}

	img, err := h.getImage(vol, uriParam.Page,
		getImageOptions{prefetch: h.config.PrefetchImages, updateCache: true})
	if err != nil {
		return nil, err
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
	vol, err := h.database.GetVolume(ctx, db.GetVolumeOptions{ID: uriParam.VolumeID})
	if err != nil {
		logger.Errorf("get volume error: %s", err)
		return nil, err
	}

	img, err := h.getImage(vol, uriParam.Page, getImageOptions{updateCache: true})
	if err != nil {
		return nil, err
	}

	// decode image for thumbnail
	image, err := imagehelper.DecodeFromBytes(img.Data)
	if err != nil {
		logger.Errorf("decode page image error: %s", err)
		return nil, err
	}
	// TODO w&h come from params
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
		Hash:   util.ImageHash(buf.Bytes()),
		Format: imageFormatJPEG,
		H:      height,
		W:      width,
	}, nil
}

func (h *handlers) apiUpdateVolumeProgress(ctx *gin.Context) (interface{}, error) {
	var req types.UpdateVolumeProgressRequest
	if err := h.bindAndValidJSON(ctx, &req); err != nil {
		return nil, err
	}

	opt := db.BatchUpdateVolumeProgressOptions{Operate: req.Operator}
	for _, vol := range req.Volumes {
		opt.SetVolumes = append(opt.SetVolumes, db.SetVolumeProgress{VolumeID: vol.ID, Page: vol.Page})
	}
	if err := h.database.BatchUpdateVolumeProgress(ctx, opt); err != nil {
		log.Errorf("batch update volume progress error: %s", err)
		return nil, err
	}

	return nil, nil
}
