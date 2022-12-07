package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chain710/manga/internal/serve"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type serveCmd struct {
	debug   bool
	baseURI string
}

const (
	addr                    = "addr"
	dsn                     = "dsn"
	archiveCache            = "archive_cache"
	pageCache               = "page_cache"
	thumbCache              = "thumb_cache"
	prefetchImage           = "prefetch_image"
	prefetchQueue           = "prefetch_queue"
	thumbWidth              = "thumb_width"
	thumbHeight             = "thumb_height"
	watchInterval           = "watch_interval"
	scanWorkerCount         = "scan_workers"
	serializeWorkerCount    = "serialize_workers"
	thumbRetryDelay         = "thumb_retry_delay"
	thumbScannerWorkerCount = "thumb_scanner_workers"
	watchDebounceInterval   = "watch_debounce_interval"
	maxDatabaseConn         = "max_db_conn"
	fullTextSearchTokenizer = "fts_tokenizer"
)

func (m *serveCmd) RunE(cmd *cobra.Command, _ []string) error {
	config := serve.Config{
		Addr:                    viper.GetString("addr"),
		Debug:                   m.debug,
		BaseURI:                 m.baseURI,
		DSN:                     viper.GetString("dsn"),
		ArchiveCacheSize:        viper.GetInt(archiveCache),
		PageCacheSize:           viper.GetInt(pageCache),
		ThumbCacheSize:          viper.GetInt(thumbCache),
		PrefetchImages:          viper.GetInt(prefetchImage),
		PrefetchQueue:           viper.GetInt(prefetchQueue),
		ThumbWidth:              viper.GetInt(thumbWidth),
		ThumbHeight:             viper.GetInt(thumbHeight),
		WatchInterval:           viper.GetDuration(watchInterval),
		ScanWorkerCount:         viper.GetInt(scanWorkerCount),
		SerializeWorkerCount:    viper.GetInt(serializeWorkerCount),
		ThumbRetryDelay:         viper.GetDuration(thumbRetryDelay),
		ThumbScannerWorkerCount: viper.GetInt(thumbScannerWorkerCount),
		WatchDebounceInterval:   viper.GetDuration(watchDebounceInterval),
		MaxDatabaseConn:         viper.GetInt(maxDatabaseConn),
		FullTextSearchTokenizer: viper.GetString(fullTextSearchTokenizer),
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
	viperFlag(realCmd.Flags(), addr, ":8080", "http serve addr")
	viperFlag(realCmd.Flags(), dsn, "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	viperFlag(realCmd.Flags(), archiveCache, 5, "archive cache size")
	viperFlag(realCmd.Flags(), pageCache, 100, "page cache size")
	viperFlag(realCmd.Flags(), thumbCache, 100, "thumb cache size")
	viperFlag(realCmd.Flags(), prefetchImage, 5, "prefetch image count. 0 means disable")
	viperFlag(realCmd.Flags(), prefetchQueue, 16, "prefetch queue")
	viperFlag(realCmd.Flags(), thumbWidth, 210, "thumb width px")
	viperFlag(realCmd.Flags(), thumbHeight, 297, "thumb height px")
	viperFlag(realCmd.Flags(), watchInterval, 24*time.Hour, "how often scan books & libraries")
	viperFlag(realCmd.Flags(), scanWorkerCount, 1, "scan worker count")
	viperFlag(realCmd.Flags(), serializeWorkerCount, 1, "serialize worker count")
	viperFlag(realCmd.Flags(), thumbRetryDelay, time.Second*30, "thumb retry delay")
	viperFlag(realCmd.Flags(), thumbScannerWorkerCount, 1, "thumb scanner worker count")
	viperFlag(realCmd.Flags(), watchDebounceInterval, time.Minute, "watch debounce interval, should >= 1s")
	viperFlag(realCmd.Flags(), maxDatabaseConn, 100, "max database connection count")
	viperFlag(realCmd.Flags(), fullTextSearchTokenizer, "default", "full text search tokenizer")
	rootCmd.AddCommand(realCmd)
}
