package cmd

import (
	"errors"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/scanner"
	"github.com/chain710/manga/internal/tasks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
)

type toolsCmd struct {
	db                db.Interface
	ignoreBookModTime bool
}

func (t *toolsCmd) setDatabase(_ *cobra.Command, _ []string) error {
	dsn := viper.GetString("dsn")
	if dsn == "" {
		return errors.New("dsn required")
	}

	database, err := db.NewPostgres(dsn, db.DefaultPostgresOptions())
	if err != nil {
		return err
	}

	t.db = database
	return nil
}

func (t *toolsCmd) addLibrary(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	lib := db.Library{
		CreateAt: db.NewTime(time.Now()),
		Name:     args[0],
		Path:     args[1],
	}

	if err := t.db.CreateLibrary(cmd.Context(), &lib); err != nil {
		log.Errorf("create library error: %s", err)
		return err
	}

	cmd.Printf("library %d created ok\n", lib.ID)
	return nil
}

func (t *toolsCmd) scanLibrary(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	cmd.SilenceUsage = true
	watcher := tasks.NewLibraryWatcher(t.db,
		tasks.WithScannerOptions(scanner.IgnoreBookModTime(t.ignoreBookModTime)))
	return watcher.Once(cmd.Context(), int64(id))
}

func (t *toolsCmd) scanThumbnail(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	cd := tasks.NewThumbnailScanner(t.db)
	return cd.Once(cmd.Context())
}

func (t *toolsCmd) setVolumeThumbnail(cmd *cobra.Command, args []string) error {
	vid, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return err
	}

	path := args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("read image file %s error: %s", path, err)
		return err
	}

	if err := t.db.SetVolumeThumbnail(cmd.Context(), db.VolumeThumbnail{ID: vid, Thumbnail: data}); err != nil {
		log.Errorf("set vol thumb error: %s", err)
		return err
	}

	cmd.Println("volume thumb set")
	return nil
}

func init() {
	var cmd toolsCmd
	addLib := &cobra.Command{
		Use:     "addlib <name> <path>",
		Args:    cobra.ExactArgs(2),
		PreRunE: cmd.setDatabase,
		RunE:    cmd.addLibrary,
	}
	scanLib := &cobra.Command{
		Use:     "scanlib <id>",
		Args:    cobra.ExactArgs(1),
		PreRunE: cmd.setDatabase,
		RunE:    cmd.scanLibrary,
	}
	scanLib.Flags().BoolVar(&cmd.ignoreBookModTime, "ignore-book-modtime", false, "ignore book last mod time")
	setVolThumb := &cobra.Command{
		Use:     "setvolthumb <vid> <image_path>",
		Args:    cobra.ExactArgs(2),
		PreRunE: cmd.setDatabase,
		RunE:    cmd.setVolumeThumbnail,
	}
	scanVolThumb := &cobra.Command{
		Use:     "scanthumb",
		PreRunE: cmd.setDatabase,
		RunE:    cmd.scanThumbnail,
	}
	realCmd := &cobra.Command{
		Use:   "tools",
		Short: "tools collection",
	}
	realCmd.AddCommand(addLib, scanLib, setVolThumb, scanVolThumb)
	rootCmd.AddCommand(realCmd)
	realCmd.PersistentFlags().StringP("dsn", "", "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	_ = viper.BindPFlag("dsn", realCmd.Flags().Lookup("dsn"))
}
