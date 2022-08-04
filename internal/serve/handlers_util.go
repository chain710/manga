package serve

import (
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/gin-gonic/gin"
)

func (h *handlers) getVolumeArchive(ctx *gin.Context, vid int64) (*db.Volume, arc.Archive, error) {
	logger := log.With("vol_id", vid)
	vol, err := h.getVolume(ctx, vid)
	if err != nil {
		logger.Errorf("get volume error: %s", err)
		return nil, nil, err
	}

	fh, err := fileHash(vol.Path)
	if err != nil {
		logger.Errorf("get volume hash error: %s", err)
		return nil, nil, err
	}
	archive, err := arc.Open(vol.Path, arc.OpenWithCache(h.archiveCache, fh),
		arc.OpenSkipReadingFiles())
	if err != nil {
		logger.Errorf("open archive %s error: %s", vol.Path, err)
		return nil, nil, err
	}

	return vol, archive, nil
}

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
