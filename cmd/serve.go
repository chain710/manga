package cmd

import (
	"context"
	"github.com/chain710/manga/internal/serve"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type serveCmd struct {
	debug   bool
	baseURI string
}

const (
	addr                 = "addr"
	dsn                  = "dsn"
	archiveCache         = "archive_cache"
	pageCache            = "page_cache"
	thumbCache           = "thumb_cache"
	prefetchImage        = "prefetch_image"
	prefetchQueue        = "prefetch_queue"
	thumbWidth           = "thumb_width"
	thumbHeight          = "thumb_height"
	watchInterval        = "watch_interval"
	scanWorkerCount      = "scan_workers"
	serializeWorkerCount = "serialize_workers"
)

func (m *serveCmd) RunE(cmd *cobra.Command, _ []string) error {
	config := serve.Config{
		Addr:                 viper.GetString("addr"),
		Debug:                m.debug,
		BaseURI:              m.baseURI,
		DSN:                  viper.GetString("dsn"),
		ArchiveCacheSize:     viper.GetInt(archiveCache),
		PageCacheSize:        viper.GetInt(pageCache),
		ThumbCacheSize:       viper.GetInt(thumbCache),
		PrefetchImages:       viper.GetInt(prefetchImage),
		PrefetchQueue:        viper.GetInt(prefetchQueue),
		ThumbWidth:           viper.GetInt(thumbWidth),
		ThumbHeight:          viper.GetInt(thumbHeight),
		WatchInterval:        viper.GetDuration(watchInterval),
		ScanWorkerCount:      viper.GetInt(scanWorkerCount),
		SerializeWorkerCount: viper.GetInt(serializeWorkerCount),
	}

	if err := config.Validate(); err != nil {
		return err
	}

	cmd.SilenceUsage = true
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
		cmd.Println("got signal, shutting down serving")
		<-quit
		cmd.Println("second signal, exit")
		os.Exit(1) // second signal. Exit directly.
	}()

	serve.Start(ctx, config)
	return nil
}

func init() {
	cmd := serveCmd{}
	realCmd := &cobra.Command{
		Use:     "serve",
		Example: "serve --addr :8080 --dsn 'postgres://localhost:5432/db?sslmode=disable'",
		RunE:    cmd.RunE,
	}

	realCmd.Flags().BoolVarP(&cmd.debug, "debug", "D", false, "debug mode")
	realCmd.Flags().StringVarP(&cmd.baseURI, "base_uri", "", "", "http base uri")
	_ = viperFlag(realCmd.Flags(), addr, ":8080", "http serve addr")
	_ = viperFlag(realCmd.Flags(), dsn, "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	_ = viperFlag(realCmd.Flags(), archiveCache, 5, "archive cache size")
	_ = viperFlag(realCmd.Flags(), pageCache, 100, "page cache size")
	_ = viperFlag(realCmd.Flags(), thumbCache, 100, "thumb cache size")
	_ = viperFlag(realCmd.Flags(), prefetchImage, 5, "prefetch image count. 0 means disable")
	_ = viperFlag(realCmd.Flags(), prefetchQueue, 16, "prefetch queue")
	_ = viperFlag(realCmd.Flags(), thumbWidth, 210, "prefetch queue")
	_ = viperFlag(realCmd.Flags(), thumbHeight, 297, "prefetch queue")
	_ = viperFlag(realCmd.Flags(), watchInterval, 24*time.Hour, "how often scan books & libraries")
	_ = viperFlag(realCmd.Flags(), scanWorkerCount, 1, "scan worker count")
	_ = viperFlag(realCmd.Flags(), serializeWorkerCount, 1, "serialize worker count")
	rootCmd.AddCommand(realCmd)
}
