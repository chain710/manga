package tasks

import (
	"context"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/parser"
	"github.com/chain710/workqueue"
	"reflect"
	"sync"
	"time"
)

func NewLibraryScanner(db db.Interface) *LibraryScanner {
	// TODO options
	interval := time.Hour * 24
	return &LibraryScanner{
		database:    db,
		interval:    interval,
		fileTypes:   []string{},
		q:           workqueue.NewRetryQueue("scanner", workqueue.NewClock()),
		workerCount: 1,
	}
}

/*
LibraryScanner scan library in database
*/
type LibraryScanner struct {
	ctx         context.Context
	database    db.Interface
	interval    time.Duration
	fileTypes   []string
	q           workqueue.RetryInterface
	workerCount int
}

func (l *LibraryScanner) Start(ctx context.Context) error {
	l.ctx = ctx
	for i := 0; i < l.workerCount; i++ {
		go l.scanBook(i)
	}

	ticker := time.NewTicker(l.interval)
	for {
		select {
		case <-ctx.Done():
			l.q.ShutDown()
			return nil
		case <-ticker.C:
			libraries, err := l.database.ListLibraries(ctx)
			if err != nil {
				log.Errorf("list library error: %s, try later", err)
				continue
			}

			for i := range libraries {
				l.scanLibrary(ctx, &libraries[i])
			}
		}
	}
}

func (l *LibraryScanner) Once(id int) error {
	l.ctx = context.TODO()
	var wg sync.WaitGroup
	wg.Add(l.workerCount)
	for i := 0; i < l.workerCount; i++ {
		go func(idx int) {
			l.scanBook(idx)
			wg.Done()
		}(i)
	}

	lib, err := l.database.GetLibrary(l.ctx, id)
	if err != nil {
		log.Errorf("get library %d error: %s", id, err)
		return err
	}

	l.scanLibrary(l.ctx, lib)
	l.q.ShutDown()
	wg.Wait()
	return nil
}

func (l *LibraryScanner) scanLibrary(ctx context.Context, lib *db.Library) {
	books, err := l.database.ListBooks(ctx, &db.ListBooksOptions{LibraryID: lib.ID})
	if err != nil {
		log.Errorf("list existing books of lib %d error %s", lib.ID, err)
		return
	}

	booksMapping := make(map[string]*db.Book)
	for i := range books {
		b := &books[i]
		booksMapping[b.Path] = b
	}

	err = parser.WalkLibrary(lib.Path,
		func(b *parser.BookMeta) {
			l.processBook(b, booksMapping[b.Path])
		},
		parser.LibraryOptionAcceptFileTypes(l.fileTypes...))
	if err != nil {
		log.Errorf("walk library %s error %s", lib.Path, err)
	} else {
		log.Infof("walk library %s ok", lib.Path)
	}
}

type bookItem struct {
	book db.Book
	new  bool // insert or update
}

func (b *bookItem) IsReplaceable() bool {
	return true
}

func (b *bookItem) Index() interface{} {
	return b.book.Path
}

func (l *LibraryScanner) processBook(book *parser.BookMeta, bookOld *db.Book) {
	// TODO: sort book files by name or number
	now := time.Now()
	var item bookItem
	if bookOld != nil {
		item.book = *bookOld
	} else {
		item.new = true
	}

	l.convertToBook(book, &item.book)
	if item.book.CreateAt.IsZero() {
		item.book.CreateAt = now
	}

	if reflect.DeepEqual(&item.book, bookOld) {
		log.Debugf("book(%d) %s unchanged", item.book.ID, item.book.Name)
		return
	}

	// update time
	item.book.UpdateAt = now
	if err := l.q.Add(&item); err != nil {
		log.Errorf("add book %s to queue error: %s", item.book.ID, item.book.Name)
	} else {
		log.Debugf("add book %s to queue", item.book.Name)
	}
}

func (l *LibraryScanner) scanBook(worker int) {
	for {
		item, shutdown := l.q.Get()
		if shutdown {
			log.Infof("book scanner %d shutdown", worker)
			return
		}

		b := item.(*bookItem)
		var err error
		if b.new {
			err = l.database.CreateBook(l.ctx, &b.book)
		} else {
			err = l.database.UpdateBook(l.ctx, &b.book)
		}
		if err != nil {
			log.Errorw("add/update book to database error", "new", b.new, "id", b.book.ID, "name", b.book.Name, "error", err)
		} else {
			log.Infow("add/update book to database ok", "new", b.new, "id", b.book.ID, "name", b.book.Name)
		}
	}
}

func (l *LibraryScanner) convertToBook(in *parser.BookMeta, out *db.Book) {
	out.Name = in.Name.Name
	out.Writer = in.Name.Writer
	out.Path = in.Path
	out.Volume = len(in.Volumes)
	out.Files = db.BookFiles{
		Volumes: l.convertToBookList(in.Volumes),
		Extras:  l.convertToBookList(in.Extras),
	}
	return
}

func (l *LibraryScanner) convertToBookList(books []parser.BookVolumeBasicMeta) []db.BookFile {
	var list []db.BookFile
	for _, volume := range books {
		list = append(list, db.BookFile{
			Name: volume.Name,
			ID:   volume.ID,
			Path: volume.Path,
		})
	}
	return list
}
