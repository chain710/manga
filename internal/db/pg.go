package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"strings"
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
	sourceDriver, err := iofs.New(migrations.FS, "pg")
	if err != nil {
		return nil, err
	}

	driver, err := pgx.WithInstance(p.DB.DB, &pgx.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
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

func (p *Postgres) PatchBook(ctx context.Context, opt PatchBookOptions) (*Book, error) {
	stmt, args := p.patchBookStmt(opt)
	stmt = p.DB.Rebind(stmt)
	rows, err := p.DB.NamedQueryContext(ctx, stmt, args)
	if err != nil {
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}

		return nil, sql.ErrNoRows
	}

	var book Book
	if err := rows.StructScan(&book); err != nil {
		return nil, err
	}

	return &book, nil
}

func (p *Postgres) DeleteBook(ctx context.Context, options DeleteBookOptions) error {
	_, err := p.DB.ExecContext(ctx, `delete from books where id=$1`, options.ID)
	return err
}

func (p *Postgres) GetLibrary(ctx context.Context, id int64) (*Library, error) {
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

func (p *Postgres) CreateLibrary(ctx context.Context, lib *Library) error {
	return p.DB.GetContext(ctx, &lib.ID, "insert into libraries (create_at, scan_at, name, path) values ($1, $2, $3, $4) returning id",
		lib.CreateAt, lib.ScanAt, lib.Name, lib.Path)
}

func (p *Postgres) DeleteLibrary(ctx context.Context, opt DeleteLibraryOptions) error {
	_, err := p.DB.ExecContext(ctx, "delete from libraries where id=$1", opt.ID)
	return err
}

func (p *Postgres) PatchLibrary(ctx context.Context, opt PatchLibraryOptions) (*Library, error) {
	var updates []string
	args := map[string]interface{}{
		"id": opt.ID,
	}
	if opt.Name != "" {
		updates = append(updates, "name=:name")
		args["name"] = opt.Name
	}
	if !opt.ScanAt.IsZero() {
		updates = append(updates, "scan_at=:scan_at")
		args["scan_at"] = opt.ScanAt
	}

	rows, err := p.DB.NamedQueryContext(ctx,
		fmt.Sprintf("update libraries set %s where id=:id returning *", strings.Join(updates, ",")),
		args)
	if err != nil {
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}

		return nil, sql.ErrNoRows
	}

	var lib Library
	if err := rows.StructScan(&lib); err != nil {
		return nil, err
	}

	return &lib, nil
}

func (p *Postgres) ListBooks(ctx context.Context, opt ListBooksOptions) ([]Book, int, error) {
	var books []bookJoin
	query, args, err := p.listBookQuery(opt)
	if err != nil {
		log.Errorf("gen list book query error: %s", err)
		return nil, 0, err
	}
	query = p.DB.Rebind(query)
	if err := p.DB.SelectContext(ctx, &books, query, args...); err != nil {
		return nil, 0, err
	}

	ret := make([]Book, len(books))
	for i := range books {
		ret[i].BookTable = books[i].BookTable
		p.setBookProgress(&ret[i], books[i].bookProgressAlias)
	}

	query, args, err = p.countBookQuery(opt)
	if err != nil {
		log.Errorf("gen count book query error: %s", err)
		return nil, 0, err
	}

	query = p.DB.Rebind(query)
	var count int
	if err := p.DB.GetContext(ctx, &count, query, args...); err != nil {
		return nil, 0, err
	}

	return ret, count, nil
}

// GetBook get book by id or path; return nil if not found
func (p *Postgres) GetBook(ctx context.Context, opt GetBookOptions) (*Book, error) {
	query, args := p.getBookQuery(opt)
	var book bookJoin
	if err := p.DB.GetContext(ctx, &book, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var ret Book
	ret.BookTable = book.BookTable
	p.setBookProgress(&ret, book.bookProgressAlias)

	var vols []Volume
	if err := p.DB.SelectContext(ctx, &vols, "select * from volumes where book_id=$1 order by volume", book.ID); err != nil {
		return nil, err
	}

	for _, vol := range vols {
		if vol.Volume > 0 {
			ret.Volumes = append(ret.Volumes, vol)
		} else {
			ret.Extras = append(ret.Extras, vol)
		}
	}
	return &ret, nil
}

func (p *Postgres) GetVolume(ctx context.Context, opt GetVolumeOptions) (*Volume, error) {
	var vol Volume
	if err := p.DB.GetContext(ctx, &vol,
		`select volumes.*,b.name as book_name, b.writer from volumes left join books b on volumes.book_id = b.id 
         where volumes.id=$1`, opt.ID); err != nil {
		return nil, err
	}

	return &vol, nil
}

func (p *Postgres) GetVolumeNeighbour(ctx context.Context, opt GetVolumeNeighbourOptions) (*int64, *int64, error) {
	prevID := new(int64)
	nextID := new(int64)

	if err := p.DB.GetContext(ctx, &nextID,
		`select id from volumes where book_id=$1 and volume > $2 and volume != 0 order by volume limit 1`,
		opt.BookID, opt.Volume); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			nextID = nil
		} else {
			return nil, nil, err
		}
	}

	if err := p.DB.GetContext(ctx, &prevID,
		`select id from volumes where book_id=$1 and volume < $2 and volume != 0 order by volume desc limit 1`,
		opt.BookID, opt.Volume); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			prevID = nil
		} else {
			return nil, nil, err
		}
	}

	return prevID, nextID, nil
}

func (p *Postgres) SetVolumeProgress(ctx context.Context, opt VolumeProgressOptions) error {
	now := clk.Now()
	_, err := p.DB.NamedExecContext(ctx, `insert into volume_progress 
    (create_at, update_at, book_id, volume_id, complete, page) 
	values (:create_at, :update_at, :book_id, :volume_id, :complete, :page) 
	on conflict(volume_id) do update 
	set update_at = :update_at, book_id = :book_id, volume_id = :volume_id,
	complete = volume_progress.complete or :complete, 
	page = :page`,
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

func (p *Postgres) BatchUpdateVolumeProgress(ctx context.Context, opt BatchUpdateVolumeProgressOptions) error {
	type insertParams struct {
		ID     int64 `db:"id"`
		BookID int64 `db:"book_id"`
	}
	switch opt.Operate {
	case UpdateVolumeProgressComplete:
		stmt, args, err := sqlx.In("select id,book_id from volumes where id in (?)", opt.IDs)
		if err != nil {
			log.Errorf("gen select in statement error: %s", err)
			return err
		}

		stmt = p.DB.Rebind(stmt)
		var filterIDs []insertParams
		if err := p.DB.SelectContext(ctx, &filterIDs, stmt, args...); err != nil {
			return err
		}

		stmt = `insert into volume_progress (create_at, update_at, book_id, volume_id, complete, page) 
values (now(),now(),:book_id,:id,true,0) on conflict (volume_id) do update set complete=true, page=0, update_at=now()`
		_, err = p.DB.NamedExecContext(ctx, stmt, filterIDs)
		return err
	case UpdateVolumeProgressReset:
		stmt, args, err := sqlx.In(`delete from volume_progress where volume_id in (?)`, opt.IDs)
		if err != nil {
			log.Errorf("gen delete in statement error: %s", err)
			return err
		}
		stmt = p.DB.Rebind(stmt)
		_, err = p.DB.ExecContext(ctx, stmt, args...)
		return err
	default:
		return errors.New("unrecognized operator")
	}
}

func (p *Postgres) SetVolumeThumbnail(ctx context.Context, thumbnail VolumeThumbnail) error {
	stmt := `insert into volume_thumbnail (id, thumbnail) values (:id, :thumbnail) 
                                                 on conflict(id) do update set thumbnail=:thumbnail`
	_, err := p.DB.NamedExecContext(ctx, stmt, map[string]interface{}{
		"id":        thumbnail.ID,
		"thumbnail": thumbnail.Thumbnail,
	})
	return err
}

func (p *Postgres) GetVolumeThumbnail(ctx context.Context, opt GetVolumeThumbOptions) (*VolumeThumbnail, error) {
	var thumbnail VolumeThumbnail
	var args []interface{}
	var stmt string
	if opt.BookID == nil {
		stmt = "select * from volume_thumbnail where id=$1"
		args = append(args, opt.ID)
	} else {
		stmt = `select distinct on (volumes.book_id) volume_thumbnail.* from volumes right join 
    volume_thumbnail on volumes.id = volume_thumbnail.id where book_id=$1 order by volumes.book_id, volumes.volume`
		args = append(args, *opt.BookID)
	}

	if err := p.DB.GetContext(ctx, &thumbnail, stmt, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &thumbnail, nil
}

func (p *Postgres) SetBookThumbnail(ctx context.Context, thumbnail BookThumbnail) error {
	stmt := `insert into book_thumbnail (id, thumbnail) values (:id, :thumbnail) 
                                                 on conflict(id) do update set thumbnail=:thumbnail`
	_, err := p.DB.NamedExecContext(ctx, stmt, map[string]interface{}{
		"id":        thumbnail.ID,
		"thumbnail": thumbnail.Thumbnail,
	})
	return err
}

func (p *Postgres) GetBookThumbnail(ctx context.Context, id int64) (*BookThumbnail, error) {
	var thumbnail BookThumbnail
	if err := p.DB.GetContext(ctx, &thumbnail, `select * from book_thumbnail where id=$1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &thumbnail, nil
}

func (p *Postgres) ListVolumes(ctx context.Context, opt ListVolumesOptions) ([]Volume, error) {
	if !opt.WithoutThumbnail {
		panic("not implemented")
	}

	var vols []Volume
	err := p.DB.SelectContext(ctx, &vols,
		`select volumes.* from volumes left join volume_thumbnail vc on volumes.id = vc.id where vc.id is null`)
	if err != nil {
		return nil, err
	}

	return vols, err
}

func (p *Postgres) replaceVolumes(ctx context.Context, tx *sqlx.Tx, volumes []Volume) error {
	var err error
	for i := range volumes {
		volume := &volumes[i]
		args := map[string]interface{}{
			"create_at":  volume.CreateAt,
			"book_id":    volume.BookID,
			"path":       volume.Path,
			"title":      volume.Title,
			"volume":     volume.Volume,
			"page_count": volume.PageCount,
			"files":      volume.Files,
			"id":         volume.ID,
		}
		var query string
		var qargs []interface{}
		if volume.ID >= 0 {
			// update
			query, qargs, err = sqlx.Named(`update volumes set
                   book_id=:book_id,
                   path=:path,
                   title=:title,
                   volume=:volume,
                   page_count=:page_count,
                   files=:files
                   where id=:id`, args)
			if err != nil {
				log.Errorf("named statement for update volume error: %s", err)
				return err
			}
			query = tx.Rebind(query)
			_, err = tx.ExecContext(ctx, query, qargs...)
		} else {
			// insert
			query, qargs, err = sqlx.Named(`insert into volumes 
    (book_id, create_at, path, title, volume, page_count, files) values 
    (:book_id, :create_at, :path, :title, :volume, :page_count, :files) returning
        id`, args)
			if err != nil {
				log.Errorf("named statement for insert volume error: %s", err)
				return err
			}
			var lastInsertID int64
			query = tx.Rebind(query)
			err = tx.GetContext(ctx, &lastInsertID, query, qargs...)
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
	var cond string
	var args []interface{}
	if opt.Path != "" {
		cond = "path = $1"
		args = append(args, opt.Path)
	} else {
		cond = "id = $1"
		args = append(args, opt.ID)
	}

	var stmt string
	if opt.WithoutProgress {
		stmt = `select * from books where ` + cond
	} else {
		stmt = `select books.*,
       b.page_count as read_volume_page_count,
       b.volume as read_volume,
       b.volume_id as read_volume_id,
       b.page as read_volume_page,
       b.title as read_volume_title,
       b.create_at as read_volume_begin_at,
       b.update_at as read_volume_at from books left join(
       select page_count, volume, title, a.*
                     from volumes
                              right join (select distinct on (book_id) book_id, page, volume_id, update_at, create_at
                                          from volume_progress
                                          order by book_id, update_at desc) a 
                                  on volumes.id = a.volume_id) b 
           on books.id=b.book_id where ` + cond
	}

	return stmt, args
}

func (p *Postgres) patchBookStmt(opt PatchBookOptions) (string, map[string]interface{}) {
	args := make(map[string]interface{})
	args["id"] = opt.ID
	var set []string
	set = append(set, "update_at=:update_at")
	args["update_at"] = clk.Now()
	if opt.Name != "" {
		set = append(set, "name=:name")
		args["name"] = opt.Name
	}

	if opt.Writer != "" {
		set = append(set, "writer=:writer")
		args["writer"] = opt.Writer
	}

	if opt.Summary != "" {
		set = append(set, "summary=:summary")
		args["summary"] = opt.Summary
	}
	stmt := fmt.Sprintf("update books set %s where id=:id returning *", strings.Join(set, ","))
	return stmt, args
}

func (p *Postgres) countBookQuery(opt ListBooksOptions) (string, []interface{}, error) {
	const selectProgress = `select volumes.book_id from volumes right join 
    (select distinct on (book_id) book_id, volume_id from volume_progress order by book_id, update_at desc
    ) a on volumes.id=a.volume_id`
	q := ""
	switch opt.Join {
	case ListBooksOnly:
		q = "select count(id) as count from books"
	case ListBookWithoutThumbnail:
		q = "select count(books.id) as count from books left join book_thumbnail bt on books.id = bt.id where bt.id is null"
	case ListBooksLeftJoinProgress:
		q = `select count(books.id) as count from books left join(` + selectProgress + `) b on books.id=b.book_id`
	case ListBooksRightJoinProgress:
		q = `select count(books.id) as count from books right join(` + selectProgress + `) b on books.id=b.book_id`
	}
	args := make(map[string]interface{})
	if opt.LibraryID != nil {
		q = q + " where library_id = :lib"
		args["lib"] = *opt.LibraryID
	}

	return sqlx.Named(q, args)
}

//goland:noinspection SqlResolve
func (p *Postgres) listBookQuery(opt ListBooksOptions) (string, []interface{}, error) {
	const selectProgress = `select page_count, volume, title, a.* from volumes right join (
    select distinct on (book_id) book_id, page, volume_id, update_at, create_at from volume_progress order by book_id, update_at desc
    ) a on volumes.id=a.volume_id `
	q := ""
	switch opt.Join {
	case ListBooksOnly:
		q = "select * from books"
	case ListBookWithoutThumbnail:
		q = "select books.* from books left join book_thumbnail bt on books.id = bt.id where bt.id is null"
	case ListBooksLeftJoinProgress:
		q = `select books.*,
       b.page_count as read_volume_page_count,
       b.volume as read_volume,
       b.volume_id as read_volume_id,
       b.page as read_volume_page,
       b.title as read_volume_title,
       b.create_at as read_volume_begin_at,
       b.update_at as read_volume_at
       from books left join(` + selectProgress + `) b on books.id=b.book_id`
	case ListBooksRightJoinProgress:
		q = `select books.*,
       b.page_count as read_volume_page_count,
       b.volume as read_volume,
       b.volume_id as read_volume_id,
       b.page as read_volume_page,
       b.title as read_volume_title,
       b.create_at as read_volume_begin_at,
       b.update_at as read_volume_at
       from books right join(` + selectProgress + `) b on books.id=b.book_id`
	}
	args := make(map[string]interface{})
	if opt.LibraryID != nil {
		q = q + " where library_id = :lib"
		args["lib"] = *opt.LibraryID
	}

	if opt.Sort != "" {
		q = q + " order by " + opt.Sort
	}

	if opt.Limit > 0 {
		q = q + " limit :limit"
		args["limit"] = opt.Limit
	}

	if opt.Offset > 0 {
		q = q + " offset :offset"
		args["offset"] = opt.Offset
	}

	return sqlx.Named(q, args)
}

func (p *Postgres) setBookProgress(b *Book, progress bookProgressAlias) {
	if progress.VolumeID == nil {
		return
	}

	b.Progress = &BookProgress{
		BookID:    b.ID,
		UpdateAt:  *progress.LastReadAt,
		Volume:    *progress.VolumeNo,
		VolumeID:  *progress.VolumeID,
		Title:     *progress.VolumeTitle,
		Page:      *progress.Page,
		PageCount: *progress.PageCount,
	}
}
