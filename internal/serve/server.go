package serve

import (
	"context"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/tasks"
	gingzip "github.com/gin-contrib/gzip"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Start(ctx context.Context, cfg Config) {
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.NewPostgres(cfg.DSN, db.DefaultPostgresOptions())
	if err != nil {
		log.Errorf("init database error: %s", err)
		return
	}
	router := gin.New()
	router.Use(
		ginzap.Ginzap(log.Logger(), time.RFC3339, false),
		gingzip.Gzip(gingzip.DefaultCompression))
	archiveCache := arc.NewArchiveCache(cfg.ArchiveCacheSize)
	volumesCache := cache.NewVolumes(cfg.VolumeCacheSize)
	imagesCache := cache.NewImages(cfg.ImageCacheSize)
	var imagePrefetch *tasks.ImagePrefetch
	if cfg.PrefetchImages > 0 {
		log.Debugf("enable image prefetch, count=%d queue=%d",
			cfg.PrefetchImages, cfg.PrefetchQueue)
		imagePrefetch = tasks.NewImagePrefetch(imagesCache, archiveCache, cfg.PrefetchQueue)
		go imagePrefetch.Start(ctx)
	}

	h := handlers{
		config:        cfg,
		database:      database,
		archiveCache:  archiveCache,
		volumesCache:  volumesCache,
		imagesCache:   imagesCache,
		imagePrefetch: imagePrefetch,
	}
	h.registerRoutes(router)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	log.Debugf("ready to launch http server: %s", cfg.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Infof("http server shutdown complete")
			} else {
				log.Errorf("http server closed unexpect: %s", err)
			}
		}
	}()

	<-ctx.Done()
	log.Infof("shutting down http server")
	err = server.Close()
	if err != nil {
		log.Errorf("http server close error: %s", err)
	}
}
