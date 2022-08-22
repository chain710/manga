package tasks

import (
	"context"
	"fmt"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/scanner"
	"github.com/chain710/workqueue"
	"time"
)

type LibraryWatcherOption func(*LibraryWatcher)

func WithSerializerWorker(count int) LibraryWatcherOption {
	return func(sd *LibraryWatcher) {
		sd.serializeWorkerCount = count
	}
}

func WithScanWorker(count int) LibraryWatcherOption {
	return func(sd *LibraryWatcher) {
		sd.scanWorkerCount = count
	}
}

func WithScannerOptions(options ...scanner.Option) LibraryWatcherOption {
	return func(sd *LibraryWatcher) {
		sd.scanOptions = options
	}
}

func WithWatchInterval(duration time.Duration) LibraryWatcherOption {
	return func(sd *LibraryWatcher) {
		sd.interval = duration
	}
}

func NewLibraryWatcher(db db.Interface, options ...LibraryWatcherOption) *LibraryWatcher {
	s := &LibraryWatcher{
		database:             db,
		interval:             time.Hour * 24,
		scanQueue:            workqueue.NewRetryQueue("scanner.scan", clk),
		serializeQueue:       workqueue.NewRetryQueue("scanner.serialize", clk),
		serializeWorkerCount: 1,
		scanWorkerCount:      1,
	}

	for _, apply := range options {
		apply(s)
	}

	return s
}

/*
LibraryWatcher watch library changes and update database
*/
type LibraryWatcher struct {
	database             db.Interface
	interval             time.Duration
	scanQueue            workqueue.RetryInterface // libraryID in queue
	serializeQueue       workqueue.RetryInterface
	serializeWorkerCount int // serialize worker count
	scanWorkerCount      int
	scanOptions          []scanner.Option
	volumesCache         *cache.Volumes
}

func (s *LibraryWatcher) Start(ctx context.Context) {
	log.Infof("start library watcher ...")
	for i := 0; i < s.serializeWorkerCount; i++ {
		go s.serialize(ctx, i)
	}

	for i := 0; i < s.scanWorkerCount; i++ {
		go s.scan(ctx, i)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.serializeQueue.ShutDown()
			s.scanQueue.ShutDown()
			return
		case <-ticker.C:
			sc := scanner.New(s.serializeQueue, s.database, s.scanOptions...)
			if err := sc.Scan(ctx); err != nil {
				log.Errorf("scan error: %s", err)
			}
		}
	}
}

// AddLibrary add library to scan queue
func (s *LibraryWatcher) AddLibrary(library db.Library) error {
	return s.scanQueue.Add(&libraryItem{library})
}

// AddBook add book to scan queue
func (s *LibraryWatcher) AddBook(book db.Book) error {
	return s.scanQueue.Add(&bookItem{book})
}

func (s *LibraryWatcher) Once(ctx context.Context, id int64) error {
	sc := scanner.New(s.serializeQueue, s.database, s.scanOptions...)
	lib, err := s.database.GetLibrary(ctx, id)
	if err != nil {
		log.Errorf("get library %d error: %s", id, err)
		return err
	}

	if err := sc.ScanLibrary(ctx, lib); err != nil {
		log.Errorf("scan %d error: %s", id, err)
		return err
	}

	s.serializeQueue.ShutDown()
	s.serialize(ctx, 0)
	return nil
}

func (s *LibraryWatcher) scan(ctx context.Context, worker int) {
	for {
		item, shutdown := s.scanQueue.Get()
		if shutdown {
			log.Infof("scanner %d shutdown", worker)
			return
		}

		switch val := item.(type) {
		case *libraryItem:
			s.scanLibrary(ctx, &val.Library)
		case *bookItem:
			s.scanBook(&val.Book)
		default:
			panic(fmt.Errorf("unknown item type: %+v", item))
		}

		s.scanQueue.Done(item, nil)
	}
}

func (s *LibraryWatcher) scanLibrary(ctx context.Context, lib *db.Library) {
	sc := scanner.New(s.serializeQueue, s.database, s.scanOptions...)
	if err := sc.ScanLibrary(ctx, lib); err != nil {
		log.Errorf("scan lib %d error: %s", lib.ID, err)
		return
	}
}

func (s *LibraryWatcher) scanBook(book *db.Book) {
	sc := scanner.New(s.serializeQueue, s.database, s.scanOptions...)
	if err := sc.ScanBook(book); err != nil {
		log.Errorf("scan book %d error: %s", book.ID, err)
		return
	}
}

func (s *LibraryWatcher) serialize(ctx context.Context, worker int) {
	for {
		item, shutdown := s.serializeQueue.Get()
		if shutdown {
			log.Infof("serializer %d shutdown", worker)
			return
		}

		b := item.(*scanner.BookItem)
		var err error
		switch b.Op {
		case scanner.OpNew:
			err = s.database.CreateBook(ctx, &b.Book)
		case scanner.OpUpdate:
			err = s.database.UpdateBook(ctx, &b.Book)
		case scanner.OpDelete:
			err = s.database.DeleteBook(ctx, db.DeleteBookOptions{ID: b.Book.ID})
		default:
			panic(fmt.Errorf("unknown op %s", b.Op))
		}

		if err != nil {
			log.Errorw("sync book in database error",
				"op", b.Op, "id", b.Book.ID, "name", b.Book.Name,
				"error", err)
		} else {
			log.Infow("op book in database ok",
				"op", b.Op, "id", b.Book.ID, "name", b.Book.Name)
		}
		// evict all volume cache
		s.evictVolumeCache(&b.Book)
	}
}

func (s *LibraryWatcher) evictVolumeCache(b *db.Book) {
	if s.volumesCache == nil {
		return
	}
	for _, vol := range b.Volumes {
		s.volumesCache.Remove(vol.ID)
	}
	for _, vol := range b.Extras {
		s.volumesCache.Remove(vol.ID)
	}
}

type libraryItem struct{ db.Library }
type libraryIndex int64

func (s *libraryItem) Index() interface{} {
	return libraryIndex(s.ID)
}

func (s *libraryItem) IsReplaceable() bool {
	return true
}

func (s *libraryItem) Equal(i interface{}) bool {
	other, ok := i.(*libraryItem)
	if !ok {
		return false
	}
	return s.ID == other.ID && s.Path == other.Path
}

type bookItem struct{ db.Book }
type bookIndex int64

func (s *bookItem) Index() interface{} {
	return bookIndex(s.ID)
}

func (s *bookItem) IsReplaceable() bool {
	return true
}

func (s *bookItem) Equal(i interface{}) bool {
	other, ok := i.(*bookItem)
	if !ok {
		return false
	}
	return s.ID == other.ID && s.Path == other.Path
}
