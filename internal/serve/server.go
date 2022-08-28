package serve

import (
	"context"
	"net/http"
	"time"

	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/tasks"
	"github.com/chain710/manga/static"
	gingzip "github.com/gin-contrib/gzip"
	ginstatic "github.com/gin-contrib/static"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
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

func startLibraryWatcher(ctx context.Context, cfg Config, database db.Interface, ts *tasks.ThumbnailScanner) *tasks.LibraryWatcher {
	var libWatcher *tasks.LibraryWatcher
	var err error
	if cfg.WatchInterval > 0 {
		libWatcher, err = tasks.NewLibraryWatcher(database, ts,
			tasks.WithWatchInterval(cfg.WatchInterval),
			tasks.WithScanWorker(cfg.ScanWorkerCount),
			tasks.WithSerializerWorker(cfg.SerializeWorkerCount))
		if err != nil {
			log.Panicf("new library watcher error: %s", err)
		}
		go libWatcher.Start(ctx)
	}
	return libWatcher
}

func startThumbScanner(ctx context.Context, database db.Interface, cfg Config, archiveCache *arc.ArchiveCache) *tasks.ThumbnailScanner {
	log.Infof("start thumb scanner")
	s := tasks.NewThumbnailScanner(database, tasks.ThumbWithArchiveCache(archiveCache),
		tasks.ThumbWithSize(cfg.ThumbWidth, cfg.ThumbHeight),
		tasks.ThumbWithRetryDelay(cfg.ThumbRetryDelay),
		tasks.ThumbScannerWorkerCount(cfg.ThumbScannerWorkerCount))
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
	log.Debugf("serve static under %s", cfg.BaseURI)
	router.Use(ginstatic.Serve(cfg.BaseURI, static.FS))
	archiveCache := arc.NewArchiveCache(cfg.ArchiveCacheSize)
	pageCache := cache.NewImages(cfg.PageCacheSize)
	thumbCache := cache.NewImages(cfg.ThumbCacheSize)
	log.Debugf("archive cache=%d, page cache=%d, thumb cache=%d",
		cfg.ArchiveCacheSize, cfg.PageCacheSize, cfg.ThumbCacheSize)
	imagePrefetch := startImagePrefetch(ctx, cfg, pageCache, archiveCache)
	thumbScanner := startThumbScanner(ctx, database, cfg, archiveCache)
	libWatcher := startLibraryWatcher(ctx, cfg, database, thumbScanner)

	h := handlers{
		config:          cfg,
		database:        database,
		archiveCache:    archiveCache,
		volumePageCache: pageCache,
		thumbCache:      thumbCache,
		imagePrefetch:   imagePrefetch,
		libWatcher:      libWatcher,
		thumbScanner:    thumbScanner,
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
