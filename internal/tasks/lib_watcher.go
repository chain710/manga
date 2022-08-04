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
		sd.workerCount = count
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
		database:    db,
		interval:    time.Hour * 24,
		q:           workqueue.NewRetryQueue("scanner", workqueue.NewClock()),
		workerCount: 1,
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
	database     db.Interface
	interval     time.Duration
	q            workqueue.RetryInterface
	workerCount  int // serialize worker count
	scanOptions  []scanner.Option
	volumesCache *cache.Volumes
}

func (s *LibraryWatcher) Start(ctx context.Context) {
	for i := 0; i < s.workerCount; i++ {
		go s.serialize(ctx, i)
	}

	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ctx.Done():
			s.q.ShutDown()
			return
		case <-ticker.C:
			sc := scanner.New(s.q, s.database, s.scanOptions...)
			if err := sc.Scan(ctx); err != nil {
				log.Errorf("scan error: %s", err)
			}
		}
	}
}

func (s *LibraryWatcher) Once(ctx context.Context, id int64) error {
	sc := scanner.New(s.q, s.database, s.scanOptions...)
	lib, err := s.database.GetLibrary(ctx, id)
	if err != nil {
		log.Errorf("get library %d error: %s", id, err)
		return err
	}

	if err := sc.ScanLibrary(ctx, lib); err != nil {
		log.Errorf("scan %d error: %s", id, err)
		return err
	}

	s.q.ShutDown()
	s.serialize(ctx, 0)
	return nil
}

func (s *LibraryWatcher) serialize(ctx context.Context, worker int) {
	for {
		item, shutdown := s.q.Get()
		if shutdown {
			log.Infof("book scanner %d shutdown", worker)
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
