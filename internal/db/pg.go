package db

import (
	"context"
	"database/sql"
	"errors"
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
	tx, err := p.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Errorf("begin tx error: %s", err)
		return err
	}

	logger := log.With("path", b.Path)
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()
	err = tx.GetContext(ctx, &lastInsertID,
		`insert into books 
    (library_id, create_at, update_at, path_mod_at, path, name, writer, volume, summary) 
	values ($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id`,
		b.LibraryID, b.CreateAt, b.UpdateAt, b.PathModAt,
		b.Path, b.Name, b.Writer,
		b.Volume, b.Summary)
	if err != nil {
		logger.Errorf("insert books error: %s", err)
		return err
	}

	b.ID = lastInsertID
	b.SyncBookID()

	if err := p.replaceVolumes(ctx, tx, b.Volumes); err != nil {
		logger.Errorf("replace volumes error: %s", err)
		return err
	}

	if err := p.replaceVolumes(ctx, tx, b.Extras); err != nil {
		logger.Errorf("replace extras error: %s", err)
		return err
	}

	return tx.Commit()
}

func (p *Postgres) UpdateBook(ctx context.Context, b *Book) error {
	tx, err := p.DB.BeginTxx(ctx, nil)
	if err != nil {
		log.Errorf("begin tx error: %s", err)
		return err
	}
	logger := log.With("path", b.Path)
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback()

	b.SyncBookID()
	_, err = tx.ExecContext(ctx,
		`update books set update_at=$1, path_mod_at=$2, name=$3, writer=$4, volume=$5, summary=$6 
             where id=$7`,
		b.UpdateAt, b.PathModAt, b.Name, b.Writer, b.Volume, b.Summary, b.ID)
	if err != nil {
		logger.Errorf("update book error: %s", err)
		return err
	}

	delStmt, args, err := sqlx.In(`delete from volumes where book_id = ? and id not in (?)`,
		b.ID, b.GetVolumeIDs())
	if err != nil {
		logger.Errorf("gen delete in statement error: %s", err)
		return err
	}
	delStmt = tx.Rebind(delStmt)
	_, err = tx.ExecContext(ctx, delStmt, args...)
	if err != nil {
		logger.Errorf("delete extra volumes error: %s", err)
		return err
	}

	if err := p.replaceVolumes(ctx, tx, b.Volumes); err != nil {
		logger.Errorf("replace volumes error: %s", err)
		return err
	}

	if err := p.replaceVolumes(ctx, tx, b.Extras); err != nil {
		logger.Errorf("replace extras error: %s", err)
		return err
	}

	return tx.Commit()
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

// ListBooks TODO offset,limit,requireVolumes
func (p *Postgres) ListBooks(ctx context.Context, _ *ListBooksOptions) ([]Book, error) {
	// TODO opt
	var books []Book
	if err := p.DB.SelectContext(ctx, &books, "select * from books"); err != nil {
		return nil, err
	}

	return books, nil
}

// GetBook get book by id or path; return nil if not found
func (p *Postgres) GetBook(ctx context.Context, opt GetBookOptions) (*Book, error) {
	query, args := p.getBookQuery(opt)
	var book Book
	if err := p.DB.GetContext(ctx, &book, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var vols []Volume
	if err := p.DB.SelectContext(ctx, &vols, "select * from volumes where book_id=$1 order by volume", book.ID); err != nil {
		return nil, err
	}

	for _, vol := range vols {
		if vol.Volume > 0 {
			book.Volumes = append(book.Volumes, vol)
		} else {
			book.Extras = append(book.Extras, vol)
		}
	}
	return &book, nil
}

func (p *Postgres) GetVolume(ctx context.Context, opt GetVolumeOptions) (*Volume, error) {
	var vol Volume
	if err := p.DB.GetContext(ctx, &vol, `select * from volumes where id=$1`, opt.ID); err != nil {
		return nil, err
	}

	return &vol, nil
}

func (p *Postgres) SetVolumeProgress(ctx context.Context, opt VolumeProgressOptions) error {
	now := clk.Now()
	_, err := p.DB.NamedExecContext(ctx, `insert into volume_progress 
    (create_at, update_at, book_id, volume_id, complete, page) 
	values (:create_at, :update_at, :book_id, :volume_id, :complete, :page) 
	on conflict(volume_id) do update 
	set update_at = :update_at, book_id = :book_id, volume_id = :volume_id,
	complete = :complete, page = :page`,
		map[string]interface{}{
			"create_at": now,
			"update_at": now,
			"book_id":   opt.BookID,
			"volume_id": opt.VolumeID,
			"complete":  opt.Complete,
			"page":      opt.Page,
		})
	return err
}

func (p *Postgres) replaceVolumes(ctx context.Context, tx *sqlx.Tx, volumes []Volume) error {
	var err error
	for i := range volumes {
		volume := &volumes[i]
		if volume.ID >= 0 {
			// update
			_, err = tx.ExecContext(ctx, `update volumes set
                   book_id=$1,
                   path=$2,
                   title=$3,
                   cover=$4,
                   volume=$5,
                   files=$6
                   where id=$7`,
				volume.BookID, volume.Path, volume.Title,
				volume.Cover, volume.Volume, volume.Files,
				volume.ID)
		} else {
			// insert
			var lastInsertID int64
			err = tx.GetContext(ctx, &lastInsertID,
				`insert into volumes 
    (book_id, create_at, path, title, cover, volume, files) values 
	($1, $2, $3, $4, $5, $6, $7) returning id`,
				volume.BookID, volume.CreateAt, volume.Path, volume.Title, volume.Cover, volume.Volume, volume.Files)
			volume.ID = lastInsertID
		}

		if err != nil {
			log.Errorf("insert/update volume %s error: %s", volume.Path, err)
			return err
		}
	}

	return nil
}

func (p *Postgres) getBookQuery(opt GetBookOptions) (string, []interface{}) {
	if opt.Path != "" {
		return "select * from books where path = $1", []interface{}{opt.Path}
	} else {
		return "select * from books where id = $1", []interface{}{opt.ID}
	}
}
