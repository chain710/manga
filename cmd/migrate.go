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
	driver         string
	dataSourceName string
}

func (m *migrateCmd) RunE(_ *cobra.Command, args []string) error {
	viper.GetInt("postgres_max_open_conns")
	viper.GetDuration("postgres_conn_max_lifetime")
	database, err := db.NewPostgres(m.dataSourceName)
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
		return migration.Up()
	case migrateDown:
		return migration.Down()
	case migrateDrop:
		return migration.Drop()
	default:
		return errors.New("invalid op")
	}
}

func init() {
	cmd := migrateCmd{driver: "pgx"}
	realCmd := &cobra.Command{
		Use:     "migrate",
		Short:   "migrate database",
		Example: "migrate up|down|drop",
		Args:    cobra.ExactArgs(1),
		RunE:    cmd.RunE,
	}

	realCmd.Flags().StringVar(&cmd.dataSourceName, "dsn", "", "data source name, like postgres://localhost:5432/db?sslmode=disable")
	_ = realCmd.MarkFlagRequired("dsn")
	rootCmd.AddCommand(realCmd)
}
