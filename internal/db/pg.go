package db

import (
	"context"
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

func (p *Postgres) CreateLibrary(ctx context.Context, l *Library) error {
	logger := log.With("name", l.Name, "path", l.Path)
	var lastInsertID int64
	err := p.DB.GetContext(ctx, &lastInsertID,
		`insert into libraries (create_at, name, path) values ($1, $2, $3) returning id`,
		l.CreateAt, l.Name, l.Path)
	if err != nil {
		return err
	} else {
		logger.Debugf("new library id=%d", l.ID)
		l.ID = lastInsertID
	}

	return nil
}

func (p *Postgres) CreateBook(ctx context.Context, b *Book) error {
	var lastInsertID int64
	err := p.DB.GetContext(ctx, &lastInsertID,
		`insert into books 
    (library_id, create_at, update_at, path, name, writer, volume, summary, files) 
values ($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id`,
		b.LibraryID, b.CreateAt, b.UpdateAt,
		b.Path, b.Name, b.Writer,
		b.Volume, b.Summary, b.Files)
	if err != nil {
		return err
	}

	b.ID = lastInsertID
	return nil
}

func (p *Postgres) UpdateBook(ctx context.Context, b *Book) error {
	_, err := p.DB.ExecContext(ctx,
		`update books set update_at=$1, name=$2, writer=$3, volume=$4, summary=$5 where id=$6`,
		b.UpdateAt, b.Name, b.Writer, b.Volume, b.Summary, b.ID)
	return err
}

func (p *Postgres) GetLibrary(ctx context.Context, id int) (*Library, error) {
	var lib Library
	if err := p.DB.GetContext(ctx, &lib, "select * from libraries where id = $1", id); err != nil {
		return nil, err
	}

	return &lib, nil
}

func (p *Postgres) ListLibraries(ctx context.Context) ([]Library, error) {
	var libs []Library
	if err := p.DB.SelectContext(ctx, &libs, "select * from libraries"); err != nil {
		return nil, err
	}

	return libs, nil
}

func (p *Postgres) ListBooks(ctx context.Context, opt *ListBooksOptions) ([]Book, error) {
	// TODO opt
	var books []Book
	if err := p.DB.SelectContext(ctx, &books, "select * from books"); err != nil {
		return nil, err
	}

	return books, nil
}
