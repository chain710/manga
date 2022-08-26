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

func startImagePrefetch(ctx context.Context, cfg Config,
	imagesCache *cache.Images,
	archiveCache *arc.ArchiveCache) *tasks.ImagePrefetch {
	var imagePrefetch *tasks.ImagePrefetch
	if cfg.PrefetchImages > 0 {
		log.Debugf("enable image prefetch, count=%d queue=%d",
			cfg.PrefetchImages, cfg.PrefetchQueue)
		imagePrefetch = tasks.NewImagePrefetch(imagesCache, archiveCache, cfg.PrefetchQueue)
		go imagePrefetch.Start(ctx)
	}

	return imagePrefetch
}

func startLibraryWatcher(ctx context.Context, cfg Config, database db.Interface) *tasks.LibraryWatcher {
	var libWatcher *tasks.LibraryWatcher
	if cfg.WatchInterval > 0 {
		libWatcher = tasks.NewLibraryWatcher(database,
			tasks.WithWatchInterval(cfg.WatchInterval),
			tasks.WithScanWorker(cfg.ScanWorkerCount),
			tasks.WithSerializerWorker(cfg.SerializeWorkerCount))
		go libWatcher.Start(ctx)
	}
	return libWatcher
}

func startThumbScanner(ctx context.Context, database db.Interface, archiveCache *arc.ArchiveCache) *tasks.ThumbnailScanner {
	s := tasks.NewThumbnailScanner(database, tasks.ThumbWithArchiveCache(archiveCache))
	go s.Start(ctx)
	return s
}

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
	pageCache := cache.NewImages(cfg.PageCacheSize)
	log.Debugf("create page cache %d", cfg.PageCacheSize)
	thumbCache := cache.NewImages(cfg.ThumbCacheSize)
	log.Debugf("create thumb cache %d", cfg.ThumbCacheSize)
	imagePrefetch := startImagePrefetch(ctx, cfg, pageCache, archiveCache)
	libWatcher := startLibraryWatcher(ctx, cfg, database)
	startThumbScanner(ctx, database, archiveCache)

	h := handlers{
		config:          cfg,
		database:        database,
		archiveCache:    archiveCache,
		volumePageCache: pageCache,
		thumbCache:      thumbCache,
		imagePrefetch:   imagePrefetch,
		libWatcher:      libWatcher,
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
				log.Panicf("http server closed unexpect: %s", err)
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
