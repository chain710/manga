package cmd

import (
	"errors"
	"github.com/chain710/manga/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	migrateUp   = "up"
	migrateDown = "down"
	migrateDrop = "drop"
)

type migrateCmd struct {
}

func (m *migrateCmd) RunE(cmd *cobra.Command, args []string) error {
	// viperBindPFlag(cmd.Flags())
	dataSourceName := viper.GetString("dsn")
	if dataSourceName == "" {
		return errors.New("dsn should not be empty")
	}

	cmd.SilenceUsage = true
	database, err := db.NewPostgres(dataSourceName, db.DefaultPostgresOptions())
	if err != nil {
		return err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer database.Close()
	migration, err := database.GetMigration()
	if err != nil {
		return err
	}
	op := args[0]
	switch op {
	case migrateUp:
		err = migration.Up()
	case migrateDown:
		err = migration.Down()
	case migrateDrop:
		err = migration.Drop()
	default:
		return errors.New("invalid op")
	}

	if err != nil {
		return err
	}

	cmd.Println("OK")
	return nil
}

func init() {
	cmd := migrateCmd{}
	realCmd := &cobra.Command{
		Use:     "migrate up|down|drop",
		Short:   "migrate database",
		Example: "migrate up",
		Args:    cobra.ExactArgs(1),
		RunE:    cmd.RunE,
	}

	viperFlag(realCmd.Flags(), "dsn", "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	rootCmd.AddCommand(realCmd)
}
