package tasks

import (
	"context"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/parser"
	"github.com/chain710/workqueue"
	"reflect"
	"sync"
	"time"
)

type LibraryScannerOption func(*LibraryScanner)

func ScanFileTypes(types ...string) LibraryScannerOption {
	return func(scanner *LibraryScanner) {
		scanner.fileTypes = types
	}
}

func ScanImageTypes(types ...string) LibraryScannerOption {
	return func(scanner *LibraryScanner) {
		scanner.imageFileTypes = types
	}
}

func ScanInterval(duration time.Duration) LibraryScannerOption {
	return func(scanner *LibraryScanner) {
		scanner.interval = duration
	}
}

func ScanIgnoreBookModTime(v bool) LibraryScannerOption {
	return func(scanner *LibraryScanner) {
		scanner.ignoreBookModTime = v
	}
}

func NewLibraryScanner(db db.Interface, options ...LibraryScannerOption) *LibraryScanner {
	s := &LibraryScanner{
		ctx:            context.TODO(),
		database:       db,
		interval:       time.Hour * 24,
		fileTypes:      []string{".zip", ".rar", ".7z"},
		imageFileTypes: []string{".bmp", ".jpg", ".png"},
		q:              workqueue.NewRetryQueue("scanner", workqueue.NewClock()),
		workerCount:    1,
	}

	for _, apply := range options {
		apply(s)
	}

	return s
}

/*
LibraryScanner scan library in database
*/
type LibraryScanner struct {
	ctx               context.Context
	database          db.Interface
	interval          time.Duration
	fileTypes         []string
	imageFileTypes    []string // image in archive
	q                 workqueue.RetryInterface
	workerCount       int
	ignoreBookModTime bool
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
				l.scanLibrary(&libraries[i])
			}
		}
	}
}

func (l *LibraryScanner) Once(ctx context.Context, id int) error {
	l.ctx = ctx
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

	l.scanLibrary(lib)
	l.q.ShutDown()
	wg.Wait()
	return nil
}

func (l *LibraryScanner) scanLibrary(lib *db.Library) {
	err := parser.WalkLibrary(lib.Path,
		func() parser.LibraryWalker {
			return l.newWalker(lib)
		},
		parser.WithAllowVolumeTypes(l.fileTypes...),
		parser.WithAllowArcFileTypes(l.imageFileTypes...))
	if err != nil {
		log.Errorf("walk library %s error %s", lib.Path, err)
	} else {
		log.Infof("walk library %s ok", lib.Path)
	}
}

func (l *LibraryScanner) newWalker(lib *db.Library) parser.LibraryWalker {
	var err error
	var book *db.Book
	return parser.LibraryWalker{
		Predict: func(b *parser.BookMeta) bool {
			book, err = l.database.GetBook(l.ctx, db.GetBookOptions{Path: b.Path})
			if err != nil {
				log.Errorf("get exist book %s error: %s", b.Path, err)
				return false
			}

			if l.ignoreBookModTime {
				return true
			}
			if book != nil && book.PathModAt.EqualTime(b.ModTime) {
				log.Debugf("path mod time unchanged since last scan: %s", b.Path)
				return false
			}

			return true
		},
		Handle: func(b *parser.BookMeta) {
			l.processBook(lib.ID, b, book)
		},
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

func (l *LibraryScanner) processBook(libraryID int64, book *parser.BookMeta, bookOld *db.Book) {
	now := time.Now()
	var item bookItem
	if bookOld != nil {
		item.book = *bookOld
	} else {
		item.new = true
		item.book.LibraryID = libraryID
	}

	l.convertToBook(book, &item.book)
	if item.book.CreateAt.IsZero() {
		item.book.CreateAt = db.NewTime(now)
	}

	if reflect.DeepEqual(&item.book, bookOld) {
		log.Debugf("book(%d) %s unchanged", item.book.ID, item.book.Name)
		return
	}

	// update time
	item.book.UpdateAt = db.NewTime(now)
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
		// TODO evict all volume cache
	}
}

func (l *LibraryScanner) convertToBook(in *parser.BookMeta, out *db.Book) {
	// known volume ids
	out.Name = in.Name.Name
	out.Writer = in.Name.Writer
	out.Path = in.Path
	out.PathModAt = db.NewTime(in.ModTime)
	out.Volume = len(in.Volumes)
	// create existing index
	indexVolumes := db.IndexVolumes(out.Volumes)
	indexExtras := db.IndexVolumes(out.Extras)
	out.Volumes = l.convertToVolumes(in.Volumes, indexVolumes)
	out.Extras = l.convertToVolumes(in.Extras, indexExtras)
	return
}

func (l *LibraryScanner) convertToVolumes(volumeMetas []parser.VolumeMeta, known map[string]db.Volume) []db.Volume {
	if len(volumeMetas) == 0 {
		return nil // for reflect.DeepEqual
	}
	vols := make([]db.Volume, len(volumeMetas))
	for i, volMeta := range volumeMetas {
		vol, ok := known[volMeta.Path]
		if !ok {
			vol.ID = -1 // TODO should create
			// NOTE: we dont know BookID yet
			vol.CreateAt = db.NewTime(clk.Now())
			vol.Path = volMeta.Path
			vol.Title = volMeta.Name
			vol.Volume = volMeta.ID
			vol.Files = convertArcFiles(volMeta.Files)
		} else {
			vol.Title = volMeta.Name
			vol.Volume = volMeta.ID
			vol.Files = convertArcFiles(volMeta.Files)
		}

		vols[i] = vol
	}
	return vols
}

func convertArcFiles(files []arc.File) []db.VolumeFile {
	vf := make([]db.VolumeFile, len(files))
	for i, f := range files {
		vf[i] = db.VolumeFile{
			Path:   f.Name(),
			Offset: f.Offset(),
			Size:   f.Size(),
		}
	}
	return vf
}
