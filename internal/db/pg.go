package db

import (
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"time"
)

type PostgresOptions struct {
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func DefaultPostgresOptions() PostgresOptions {
	return PostgresOptions{
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Second,
	}
}

func NewPostgres(dataSourceName string, opt PostgresOptions) (*Postgres, error) {
	db, err := sqlx.Connect("pgx", dataSourceName)
	if err != nil {
		log.Errorf("connect pgx database error: %s", err)
		return nil, err
	}
	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)
	log.Debugf("success open pgx %s", dataSourceName)
	return &Postgres{DB: *db}, nil
}

type Postgres struct {
	sqlx.DB
}

var _ Interface = &Postgres{}

func (p *Postgres) GetMigration() (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrations.FS, "pq")
	if err != nil {
		return nil, err
	}

	driver, err := pgx.WithInstance(p.DB.DB, &pgx.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
}
