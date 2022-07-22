package db

import "github.com/golang-migrate/migrate/v4"

type Interface interface {
	// GetMigration for migrate
	GetMigration() (*migrate.Migrate, error)
}
