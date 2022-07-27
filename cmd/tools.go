package cmd

import (
	"errors"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/tasks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

type toolsCmd struct {
	db db.Interface
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
		CreateAt: time.Now(),
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
	scanner := tasks.NewLibraryScanner(t.db)
	return scanner.Once(id)
}

func init() {
	var cmd toolsCmd
	realCmd := &cobra.Command{
		Use:   "tools",
		Short: "tools collection",
	}
	realCmd.AddCommand(
		&cobra.Command{
			Use:     "addlib <name> <path>",
			Args:    cobra.ExactArgs(2),
			PreRunE: cmd.setDatabase,
			RunE:    cmd.addLibrary,
		},
		&cobra.Command{
			Use:     "scanlib <id>",
			Args:    cobra.ExactArgs(1),
			PreRunE: cmd.setDatabase,
			RunE:    cmd.scanLibrary,
		})
	rootCmd.AddCommand(realCmd)
	realCmd.PersistentFlags().StringP("dsn", "", "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	_ = viper.BindPFlag("dsn", realCmd.Flags().Lookup("dsn"))
}
