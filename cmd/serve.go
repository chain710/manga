package cmd

import (
	"context"
	"github.com/chain710/manga/internal/serve"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

type serveCmd struct {
	debug   bool
	baseURI string
}

func (m *serveCmd) RunE(cmd *cobra.Command, _ []string) error {
	config := serve.Config{
		Addr:             viper.GetString("addr"),
		Debug:            m.debug,
		BaseURI:          m.baseURI,
		DSN:              viper.GetString("dsn"),
		ArchiveCacheSize: viper.GetInt("archive_cache"),
		VolumeCacheSize:  viper.GetInt("volume_cache"),
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

	realCmd.Flags().StringP("addr", "", ":8080", "http serve addr")
	_ = viper.BindPFlag("addr", realCmd.Flags().Lookup("addr"))
	realCmd.Flags().StringP("dsn", "", "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	_ = viper.BindPFlag("dsn", realCmd.Flags().Lookup("dsn"))
	realCmd.Flags().BoolVarP(&cmd.debug, "debug", "D", false, "debug mode")
	realCmd.Flags().StringVarP(&cmd.baseURI, "base_uri", "", "", "http base uri")
	_ = viperFlag(realCmd.Flags(), "archive_cache", 100, "archive cache size")
	_ = viperFlag(realCmd.Flags(), "volume_cache", 100, "volume cache size")
	rootCmd.AddCommand(realCmd)
}
