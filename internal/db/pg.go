package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/migrations"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
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
		log.Errorf("connect pgx database error: %s; source: %s", err, dataSourceName)
		return nil, err
	}
	db.SetMaxOpenConns(opt.MaxOpenConns)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)
	log.Debugf("success open pgx")
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

	driver, err := migratepgx.WithInstance(p.DB.DB, &migratepgx.Config{})
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

	volumeIDS := b.GetVolumeIDs()
	if len(volumeIDS) > 0 {
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
	if err := p.DB.SelectContext(ctx, &libs, "select * from libraries order by id"); err != nil {
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
		log.Errorf("select books error: %s; query %s", err, query)
		return nil, 0, err
	}

	ret := make([]Book, len(books))
	for i := range books {
		b := books[i].convert(nil)
		ret[i] = *b
	}

	query, args, err = p.countBookQuery(opt)
	if err != nil {
		log.Errorf("gen count book query error: %s", err)
		return nil, 0, err
	}

	query = p.DB.Rebind(query)
	var count int
	if err := p.DB.GetContext(ctx, &count, query, args...); err != nil {
		log.Errorf("count books error: %s; query %s", err, query)
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

	listVol := ListVolumesOptions{
		BookID: &book.ID,
	}
	if GetBookJoinVolumeProgress == opt.Join {
		listVol.Join = VolumeCompactProgress
	}
	stmt, args := p.listVolumeQuery(listVol)
	var vols []volumeJoin
	if err := p.DB.SelectContext(ctx, &vols, stmt, args...); err != nil {
		return nil, err
	}
	ret := book.convert(vols)
	return ret, nil
}

func (p *Postgres) GetVolume(ctx context.Context, opt GetVolumeOptions) (*Volume, error) {
	var vol volumeJoin
	stmt, args := p.getVolumeQuery(opt)
	if err := p.DB.GetContext(ctx, &vol, stmt, args...); err != nil {
		return nil, err
	}

	return vol.convert(), nil
}

func (p *Postgres) GetVolumeNeighbour(ctx context.Context, opt GetVolumeNeighbourOptions) (*int64, *int64, error) {
	prevID := new(int64)
	nextID := new(int64)

	if opt.Volume > 0 {
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
	} else {
		if err := p.DB.GetContext(ctx, &nextID,
			`select id from volumes where book_id=$1 and id > $2 and volume = 0 order by id limit 1`,
			opt.BookID, opt.VolumeID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				nextID = nil
			} else {
				return nil, nil, err
			}
		}

		if err := p.DB.GetContext(ctx, &prevID,
			`select id from volumes where book_id=$1 and id < $2 and volume = 0 order by id desc limit 1`,
			opt.BookID, opt.VolumeID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				prevID = nil
			} else {
				return nil, nil, err
			}
		}
	}

	return prevID, nextID, nil
}

func (p *Postgres) BatchUpdateVolumeProgress(ctx context.Context, opt BatchUpdateVolumeProgressOptions) error {
	type insertParams struct {
		ID     int64 `db:"id"`
		BookID int64 `db:"book_id"`
	}
	type updateParams struct {
		insertParams
		PageCount int `db:"page_count"`
	}

	switch opt.Operate {
	case UpdateVolumeProgressUpdate:
		if len(opt.SetVolumes) != 1 {
			panic("update progress only support one")
		}

		v := opt.SetVolumes[0]
		var update updateParams
		if err := p.DB.GetContext(ctx, &update, `select id,book_id,page_count from volumes where id=$1`, v.VolumeID); err != nil {
			return err
		}

		stmt := `insert into volume_progress (create_at, update_at, book_id, volume_id, complete, page) 
values (now(),now(),$1,$2,$3::int=$4::int,$5) 
on conflict (volume_id) 
    do update set complete=($6::int=$7::int), page=$8, update_at=now()`
		// NOTE can not use named exec, see: https://github.com/jmoiron/sqlx/issues/193
		_, err := p.DB.ExecContext(ctx, stmt,
			update.BookID,
			v.VolumeID,
			v.Page,
			update.PageCount,
			v.Page,
			v.Page,
			update.PageCount,
			v.Page)
		return err
	case UpdateVolumeProgressComplete:
		stmt, args, err := sqlx.In("select id,book_id from volumes where id in (?)", opt.IDs())
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
		stmt, args, err := sqlx.In(`delete from volume_progress where volume_id in (?)`, opt.IDs())
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
	stmt := `insert into volume_thumbnail (id, hash, thumbnail) values (:id, :hash, :thumbnail) 
                                                 on conflict(id) do update set hash=:hash, thumbnail=:thumbnail`
	_, err := p.DB.NamedExecContext(ctx, stmt, map[string]interface{}{
		"id":        thumbnail.ID,
		"hash":      thumbnail.Hash,
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
    volume_thumbnail on volumes.id = volume_thumbnail.id where book_id=$1 and volumes.volume != 0 order by volumes.book_id, volumes.volume`
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
	stmt := `insert into book_thumbnail (id, hash, thumbnail) values (:id, :hash, :thumbnail) 
                                                 on conflict(id) do update set hash=:hash, thumbnail=:thumbnail`
	_, err := p.DB.NamedExecContext(ctx, stmt, map[string]interface{}{
		"id":        thumbnail.ID,
		"hash":      thumbnail.Hash,
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
	stmt, args := p.listVolumeQuery(opt)
	var vols []volumeJoin
	err := p.DB.SelectContext(ctx, &vols, stmt, args...)
	if err != nil {
		return nil, err
	}

	ret := make([]Volume, len(vols))
	for i := range ret {
		ret[i] = *vols[i].convert()
	}

	return ret, err
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
	var stmt []string
	var args []interface{}
	stmt = append(stmt, `select * from books`)
	if opt.Path != "" {
		stmt = append(stmt, "where path = $1")
		args = append(args, opt.Path)
	} else {
		stmt = append(stmt, "where id = $1")
		args = append(args, opt.ID)
	}

	return strings.Join(stmt, " "), args
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
	var where []string
	switch opt.Join {
	case ListBooksOnly:
		q = "select count(id) as count from books"
	case ListBookWithoutThumbnail:
		q = "select count(books.id) as count from books left join book_thumbnail bt on books.id = bt.id"
		where = append(where, "bt.id is null")
	case ListBooksLeftJoinProgress:
		q = `select count(id) as count from books`
	case ListBooksRightJoinProgress:
		q = `select count(books.id) as count from books right join(` + selectProgress + `) b on books.id=b.book_id`
	}
	args := make(map[string]interface{})
	if opt.LibraryID != nil {
		if len(where) > 0 {
			where = append(where, "and library_id = :lib")
		} else {
			where = append(where, "library_id = :lib")
		}
		args["lib"] = *opt.LibraryID
	}

	final := []string{q}
	if len(where) > 0 {
		final = append(final, "where")
		final = append(final, where...)
	}

	return sqlx.Named(strings.Join(final, " "), args)
}

//goland:noinspection SqlResolve
func (p *Postgres) listBookQuery(opt ListBooksOptions) (string, []interface{}, error) {
	const selectProgress = `select page_count, volume, title, a.* from volumes right join (
    select distinct on (book_id) book_id, page, volume_id, update_at, create_at from volume_progress order by book_id, update_at desc
    ) a on volumes.id=a.volume_id `
	sb := newSQLBuilder(logicOpAnd)

	switch opt.Join {
	case ListBooksOnly:
		sb.Statement("select * from books")
	case ListBookWithoutThumbnail:
		sb.Statement("select books.* from books left join book_thumbnail bt on books.id = bt.id")
		sb.Filter("bt.id is null")
	case ListBooksLeftJoinProgress:
		sb.Statement(`select books.*,
       b.page_count as read_volume_page_count,
       b.volume as read_volume,
       b.volume_id as read_volume_id,
       b.page as read_volume_page,
       b.title as read_volume_title,
       b.create_at as read_volume_begin_at,
       b.update_at as read_volume_at
       from books`).Statement(`left join (` + selectProgress + `) b on books.id=b.book_id`)
	case ListBooksRightJoinProgress:
		sb.Statement(`select books.*,
       b.page_count as read_volume_page_count,
       b.volume as read_volume,
       b.volume_id as read_volume_id,
       b.page as read_volume_page,
       b.title as read_volume_title,
       b.create_at as read_volume_begin_at,
       b.update_at as read_volume_at
       from books`).Statement(`right join(` + selectProgress + `) b on books.id=b.book_id`)
	}
	args := make(map[string]interface{})
	if opt.LibraryID != nil {
		sb.Filter("library_id = :lib")
		args["lib"] = *opt.LibraryID
	}

	if opt.Sort != "" {
		sb.Order("order by " + opt.Sort)
	} else {
		sb.Order("order by books.id")
	}

	if opt.Limit > 0 {
		sb.Order(fmt.Sprintf("limit %d", opt.Limit))
	}

	if opt.Offset > 0 {
		sb.Order(fmt.Sprintf("offset %d", opt.Offset))
	}

	q := sb.ToSQL()
	return sqlx.Named(q, args)
}

func (p *Postgres) getVolumeQuery(opt GetVolumeOptions) (string, []interface{}) {
	var stmt string
	var args []interface{}
	switch opt.Join {
	case VolumeLeftJoinBookProgress: // with book and progress
		stmt = `select volumes.*,b.name as book_name, b.writer,
       vp.volume_id as read_volume_id,
       volumes.volume as read_volume,
       volumes.title as read_volume_title,
       vp.create_at as read_volume_begin_at,
       vp.update_at as read_volume_at,
       vp.page as read_volume_page,
       volumes.page_count as read_volume_page_count
from volumes left join books b on volumes.book_id = b.id left join volume_progress vp on volumes.id = vp.volume_id
where volumes.id=$1`
		args = append(args, opt.ID)
	case "":
		stmt = `select volumes.*,b.name as book_name, b.writer from volumes left join books b on volumes.book_id = b.id 
         where volumes.id=$1`
		args = append(args, opt.ID)
	default:
		panic(fmt.Errorf("invalid join %s", opt.Join))
	}
	return stmt, args
}

func (p *Postgres) listVolumeQuery(opt ListVolumesOptions) (string, []interface{}) {
	var args []interface{}
	sb := newSQLBuilder(logicOpAnd)
	switch opt.Join {
	case VolumeCompactProgress: // with progress
		sb.Statement(`select volumes.id, 
       volumes.book_id, volumes.create_at, volumes.path, volumes.title, volumes.volume, volumes.page_count, 
       vp.volume_id as read_volume_id,
       volumes.volume as read_volume,
       volumes.title as read_volume_title,
       vp.create_at as read_volume_begin_at,
       vp.update_at as read_volume_at,
       vp.page as read_volume_page,
       volumes.page_count as read_volume_page_count
from volumes left join volume_progress vp on volumes.id = vp.volume_id`)
		if opt.BookID != nil {
			sb.Filter(`volumes.book_id=$1`)
			sb.Order(`order by volume, id`)
			args = append(args, *opt.BookID)
		}
	case VolumeReading:
		sb.Statement(`select volumes.id, 
       volumes.book_id, volumes.create_at, volumes.path, volumes.title, volumes.volume, volumes.page_count, 
       b.name as book_name, b.writer,
       vp.volume_id as read_volume_id,
       volumes.volume as read_volume,
       volumes.title as read_volume_title,
       vp.create_at as read_volume_begin_at,
       vp.update_at as read_volume_at,
       vp.page as read_volume_page,
       volumes.page_count as read_volume_page_count
from volumes left join books b on volumes.book_id = b.id inner join volume_progress vp on volumes.id = vp.volume_id and vp.complete = false`)
		if opt.BookID != nil {
			sb.Filter(`volumes.book_id=$1`)
			args = append(args, *opt.BookID)
		}

		sb.Order(`order by read_volume_at desc,volume, id asc`)
	case VolumeMustNotHaveThumb:
		sb.Statement(`select volumes.id, 
       volumes.book_id, volumes.create_at, volumes.path, volumes.title, volumes.volume, volumes.page_count, volumes.files
from volumes left join volume_thumbnail vc on volumes.id = vc.id left join books on books.id = volumes.book_id`)
		sb.Filter("vc.id is null")
		if opt.BookID != nil {
			sb.Filter("volumes.book_id=$1")
			args = append(args, *opt.BookID)
		} else if opt.LibraryID != nil {
			sb.Filter("books.library_id=$1")
			args = append(args, *opt.LibraryID)
		}
	case "":
		sb.Statement(`select * from volumes`)
		if opt.BookID != nil {
			sb.Filter(`book_id=$1`)
			args = append(args, *opt.BookID)
		}
	default:
		panic(fmt.Errorf("invalid join %s", opt.Join))
	}

	if opt.Limit > 0 {
		sb.Order(fmt.Sprintf("limit %d", opt.Limit))
	}
	return sb.ToSQL(), args
}
