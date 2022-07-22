package db

import (
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	"time"
)

func NewPostgres(dataSourceName string) (*Postgres, error) {
	db, err := sqlx.Connect("pgx", dataSourceName)
	if err != nil {
		log.Errorf("connect pgx database error: %s", err)
		return nil, err
	}
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Minute)
	return &Postgres{DB: *db}, nil
}

type Postgres struct {
	sqlx.DB
}

var _ Interface = &Postgres{}

func (p *Postgres) GetMigration() (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrations.FS, "migrations/pq")
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(p.DB.DB, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
}
