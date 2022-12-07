package cmd

import (
	"errors"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/scanner"
	"github.com/chain710/manga/internal/tasks"
	"github.com/chain710/manga/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
)

type toolsCmd struct {
	db                db.Interface
	ignoreBookModTime bool
	forceUpdateThumb  bool
}

func (t *toolsCmd) initCmd(_ *cobra.Command, _ []string) error {
	dsnValue := viper.GetString(dsn)
	if dsnValue == "" {
		return errors.New("dsn required")
	}

	options := db.DefaultPostgresOptions()
	options.Tokenizer = viper.GetString(fullTextSearchTokenizer)
	database, err := db.NewPostgres(dsnValue, options)
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
	watcher, err := tasks.NewLibraryWatcher(t.db, nil,
		tasks.WithScannerOptions(scanner.IgnoreBookModTime(t.ignoreBookModTime)))
	if err != nil {
		return err
	}
	return watcher.Once(cmd.Context(), int64(id))
}

func (t *toolsCmd) scanThumbnail(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	cd := tasks.NewThumbnailScanner(t.db)
	return cd.Once(cmd.Context(), t.forceUpdateThumb)
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

	vt := db.VolumeThumbnail{ID: vid, Hash: util.ImageHash(data), Thumbnail: data}
	if err := t.db.SetVolumeThumbnail(cmd.Context(), vt); err != nil {
		log.Errorf("set vol thumb error: %s", err)
		return err
	}

	cmd.Println("volume thumb set")
	return nil
}

func (t *toolsCmd) searchBooks(cmd *cobra.Command, args []string) error {
	query := args[0]
	books, _, err := t.db.ListBooks(cmd.Context(), db.ListBooksOptions{Query: query, Join: db.ListBooksOnly})
	if err != nil {
		log.Errorf("search books error: %s", err)
		return err
	}

	cmd.Printf("got %d results\n", len(books))
	for _, book := range books {
		cmd.Printf("%d %s (by %s) %s\n",
			book.ID, book.Name, book.Writer, book.Path)
	}
	return nil
}

func init() {
	var cmd toolsCmd
	addLib := &cobra.Command{
		Use:     "addlib <name> <path>",
		Args:    cobra.ExactArgs(2),
		PreRunE: cmd.initCmd,
		RunE:    cmd.addLibrary,
	}
	scanLib := &cobra.Command{
		Use:     "scanlib <id>",
		Args:    cobra.ExactArgs(1),
		PreRunE: cmd.initCmd,
		RunE:    cmd.scanLibrary,
	}
	scanLib.Flags().BoolVar(&cmd.ignoreBookModTime, "ignore-book-modtime", false, "ignore book last mod time")
	setVolThumb := &cobra.Command{
		Use:     "setvolthumb <vid> <image_path>",
		Args:    cobra.ExactArgs(2),
		PreRunE: cmd.initCmd,
		RunE:    cmd.setVolumeThumbnail,
	}
	scanThumb := &cobra.Command{
		Use:     "scanthumb",
		PreRunE: cmd.initCmd,
		RunE:    cmd.scanThumbnail,
	}
	scanThumb.Flags().BoolVar(&cmd.forceUpdateThumb, "force", false, "force update all thumbs")
	searchBooks := &cobra.Command{
		Use:     "searchbooks",
		PreRunE: cmd.initCmd,
		RunE:    cmd.searchBooks,
		Args:    cobra.ExactArgs(1),
	}
	realCmd := &cobra.Command{
		Use:   "tools",
		Short: "tools collection",
	}
	realCmd.AddCommand(addLib, scanLib, setVolThumb, scanThumb, searchBooks)
	viperFlag(realCmd.PersistentFlags(), dsn, "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	viperFlag(realCmd.PersistentFlags(), fullTextSearchTokenizer, "simple", "full text search tokenizer")
	rootCmd.AddCommand(realCmd)
}
