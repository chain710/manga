package scanner

import (
	"context"
	"errors"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	internalstrings "github.com/chain710/manga/internal/strings"
	"github.com/chain710/workqueue"
	"github.com/hashicorp/go-multierror"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

//goland:noinspection RegExpRedundantEscape
var (
	regexDelimiter = regexp.MustCompile(`[^\s\]\[_\-\.]+`)
)

func New(q workqueue.RetryInterface, database db.Interface, options ...Option) *Type {
	t := &Type{
		db:                 database,
		allowVolumeTypes:   internalstrings.NewSet(strings.ToLower, ".zip", ".rar", ".7z"),
		allowHiddenVolume:  false,
		allowHiddenArcFile: false,
		allowArcFileTypes:  internalstrings.NewSet(strings.ToLower, ".bmp", ".jpg", ".png"),
		sortArchiveFiles: func(files []arc.File) {
			SortSliceByDigit(files, extractArchiveFileDigits(files))
		},
		sortVolumes: func(metas []volumeMeta) {
			SortSliceByDigit(metas, extractVolumeMetaDigits(metas))
		},
		ignoreBookModTime: false,
		q:                 q,
	}

	for _, apply := range options {
		apply(t)
	}

	return t
}

type Type struct {
	db                 db.Interface
	allowVolumeTypes   internalstrings.Set
	allowHiddenVolume  bool
	allowHiddenArcFile bool
	allowArcFileTypes  internalstrings.Set
	sortArchiveFiles   func([]arc.File)
	sortVolumes        func([]volumeMeta)
	ignoreBookModTime  bool
	q                  workqueue.RetryInterface
}

// Scan db.Book to queue
func (l *Type) Scan(ctx context.Context) error {
	libraries, err := l.db.ListLibraries(ctx)
	if err != nil {
		log.Errorf("list library error: %s, try later", err)
		return err
	}

	var merr *multierror.Error
	for i := range libraries {
		lib := &libraries[i]
		merr = multierror.Append(merr, l.ScanLibrary(ctx, lib))
		// update last scan at
		patchOption := db.PatchLibraryOptions{ID: lib.ID, ScanAt: db.NewTime(clk.Now())}
		if _, err := l.db.PatchLibrary(ctx, patchOption); err != nil {
			log.Errorf("update lib %d last scan time error: %s", lib.ID, err)
		}
	}
	return merr.ErrorOrNil()
}

func (l *Type) parseVolume(path string) (*volumeMeta, error) {
	logger := log.With("vol", path)
	basic, err := parseVolumeBasic(path)
	if err != nil {
		logger.Errorf("parse vol basic meta error: %s", err)
		return nil, err
	}
	archive, err := arc.Open(path)
	if err != nil {
		logger.Errorf("open volume archive error %s", err)
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer archive.Close()
	files := archive.GetFiles()
	filesInVol := make([]arc.File, 0, len(files))
	for _, file := range files {
		if l.isArcFileAllowed(file.Name()) {
			filesInVol = append(filesInVol, file)
		}
	}

	if l.sortArchiveFiles != nil {
		l.sortArchiveFiles(filesInVol)
	}

	return &volumeMeta{
		volumeBasicMeta: basic,
		Files:           filesInVol,
	}, nil
}

func (l *Type) isArcFileAllowed(path string) bool {
	base := filepath.Base(path)
	hidden := base[0:1] == "."
	if !l.allowHiddenArcFile && hidden {
		return false
	}
	ext := filepath.Ext(path)
	return l.allowArcFileTypes.Len() == 0 || l.allowArcFileTypes.Contains(ext)
}

type classifiedEntries struct {
	volumes     []fs.DirEntry // all should be files
	directories []fs.DirEntry // other directories
}

// classifyFiles return accept files and not accept files, hidden files are dropped
func (l *Type) classifyFiles(root string, files []fs.DirEntry) classifiedEntries {
	var ce classifiedEntries
	for i := range files {
		file := files[i]
		path := filepath.Join(root, file.Name())
		if !l.allowHiddenVolume {
			if isHidden, err := isHiddenFile(path); err != nil || isHidden {
				if err != nil {
					log.Warnf("determine file %s hidden error: %s", path, err)
				}
				continue
			}
		}

		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (l.allowVolumeTypes.Len() == 0 || l.allowVolumeTypes.Contains(ext)) {
			ce.volumes = append(ce.volumes, file)
		} else if file.IsDir() {
			ce.directories = append(ce.directories, file)
		}
	}

	return ce
}

func (l *Type) getKnownBooks(ctx context.Context, libraryID int64) (map[string]db.Book, error) {
	bookMap := make(map[string]db.Book)
	books, _, err := l.db.ListBooks(ctx, db.ListBooksOptions{LibraryID: &libraryID, Join: db.ListBooksOnly})
	if err != nil {
		return nil, err
	}
	for i := range books {
		bookMap[books[i].Path] = books[i]
	}

	return bookMap, nil
}

func (l *Type) ScanLibrary(ctx context.Context, lib *db.Library) error {
	knownBooks, err := l.getKnownBooks(ctx, lib.ID)
	if err != nil {
		log.Errorf("get known books of lib %d error: %s", lib.ID, err)
		return err
	}
	info, err := os.Lstat(lib.Path)
	if err != nil {
		log.Errorf("stat path %s error: %s", lib.Path, err)
		return err
	} else if !info.IsDir() {
		return errors.New("not directory")
	}

	if err := l.walkDir(ctx, lib, lib.Path, info, knownBooks); err != nil {
		return err
	}

	// remaining knownBooks should be deleted
	for _, book := range knownBooks {
		var item BookItem
		item.Op = OpDelete
		item.Book = book
		if err := l.q.Add(&BookItem{Op: OpDelete, Book: book}); err != nil {
			log.Errorf("delete not exist book %d:%s, but add queue error: %s",
				book.ID, book.Path, err)
		} else {
			log.Infof("delete not exist book %d:%s", book.ID, book.Path)
		}
	}

	return nil
}

func (l *Type) getScanBookOptions(b *db.Book, options ...ScanBookOption) (*ScanBookOptions, error) {
	var option ScanBookOptions
	for _, apply := range options {
		apply(&option)
	}
	if b != nil {
		option.libraryID = b.LibraryID
	}
	if b != nil {
		option.path = b.Path
	}
	if option.modTime.IsZero() {
		info, err := os.Lstat(option.path)
		if err != nil {
			log.Errorf("stat book path %s error: %s", option.path, err)
			return nil, err
		}
		option.modTime = info.ModTime()
	}

	if option.entries == nil {
		entries, err := os.ReadDir(option.path)
		if err != nil {
			log.Errorf("readdir path %s error: %s", option.path, err)
			return nil, err
		}

		cf := l.classifyFiles(option.path, entries)
		option.entries = &cf
	}

	return &option, nil
}

func (l *Type) ScanBook(b *db.Book, options ...ScanBookOption) error {
	option, err := l.getScanBookOptions(b, options...)
	if err != nil {
		log.Errorf("get scan book options error: %s", err)
		return err
	}

	logger := log.With("book_path", option.path)
	book := bookMeta{
		Path: option.path,
	}

	book.bookNameMeta = parseBookName(filepath.Base(option.path))
	if l.shouldSkipBook(option.modTime, b) {
		return nil
	}

	for _, f := range option.entries.volumes {
		volPath := filepath.Join(option.path, f.Name())
		vol, err := l.parseVolume(volPath)
		if err != nil {
			logger.Errorf("parse vol %s error: %s", volPath, err)
			continue // parse remaining files
		}

		if len(vol.Files) == 0 {
			logger.Debugf("volume file is empty: %s", volPath)
			continue
		}

		book.Volumes = append(book.Volumes, *vol)
	}

	if l.sortVolumes != nil {
		l.sortVolumes(book.Volumes)
	}

	// assign sequential volume id here
	for i := range book.Volumes {
		book.Volumes[i].ID = i + 1
	}

	for i := range option.entries.directories {
		entry := option.entries.directories[i]
		entryPath := filepath.Join(option.path, entry.Name())
		if err := l.walkExtra(entryPath, &book); err != nil {
			return err
		}
	}

	if l.sortVolumes != nil {
		l.sortVolumes(book.Extras)
	}

	l.handleBook(option.libraryID, &book, b)
	return nil
}

func (l *Type) shouldSkipBook(modTime time.Time, book *db.Book) bool {
	if l.ignoreBookModTime || book == nil {
		return false
	}

	if book.PathModAt.EqualTime(modTime) {
		log.Debugf("path mod time unchanged since last scan: %s", book.Path)
		return true
	}

	return false
}

func (l *Type) handleDirs(ctx context.Context, lib *db.Library, root string, entries []fs.DirEntry, knownBooks map[string]db.Book) error {
	for _, entry := range entries {
		entryPath := filepath.Join(root, entry.Name())
		dirInfo, err := os.Lstat(entryPath)
		if err != nil {
			return err
		}
		if err := l.walkDir(ctx, lib, entryPath, dirInfo, knownBooks); err != nil {
			return err
		}
	}

	return nil
}

func (l *Type) walkDir(ctx context.Context, lib *db.Library, root string,
	info os.FileInfo,
	knownBooks map[string]db.Book) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	cf := l.classifyFiles(root, entries)
	if len(cf.volumes) == 0 {
		return l.handleDirs(ctx, lib, root, cf.directories, knownBooks)
	}

	bookInDatabase, err := l.db.GetBook(ctx, db.GetBookOptions{Path: root, WithoutProgress: true})
	if err != nil {
		log.Errorf("get exist book by path %s error: %s", root, err)
		return err
	}

	delete(knownBooks, root)
	if err := l.ScanBook(bookInDatabase, scanBookOptions(ScanBookOptions{
		libraryID: lib.ID,
		path:      root,
		modTime:   info.ModTime(),
		entries:   &cf,
	})); err != nil {
		log.Errorf("scan book by path %s error: %s", root, err)
		return err
	}

	return nil
}

func (l *Type) handleBook(libraryID int64, book *bookMeta, bookOld *db.Book) {
	now := clk.Now()
	var item BookItem
	if bookOld != nil {
		item.Op = OpUpdate
		item.Book = *bookOld
	} else {
		item.Op = OpNew
		item.Book.LibraryID = libraryID
	}

	l.convertToBook(book, &item.Book)
	if item.Book.CreateAt.IsZero() {
		item.Book.CreateAt = db.NewTime(now)
	}

	if reflect.DeepEqual(&item.Book, bookOld) {
		log.Debugf("book(%d) %s unchanged", item.Book.ID, item.Book.Name)
		return
	}

	// update time
	item.Book.UpdateAt = db.NewTime(now)
	if err := l.q.Add(&item); err != nil {
		log.Errorf("add book %s to queue error: %s", item.Book.ID, item.Book.Name)
	} else {
		log.Debugf("add book %s to queue", item.Book.Name)
	}
}

func (l *Type) walkExtra(root string, book *bookMeta) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	cf := l.classifyFiles(root, entries)
	for i := range cf.volumes {
		// add to book extra
		volPath := filepath.Join(root, cf.volumes[i].Name())
		vol, err := l.parseVolume(volPath)
		if err != nil {
			log.Errorf("parse extra vol %s error: %s", volPath, err)
			continue
		}

		if len(vol.Files) == 0 {
			log.Debugf("extra file is empty: %s", volPath)
			continue
		}

		book.Extras = append(book.Extras, *vol)
	}

	for i := range cf.directories {
		entry := cf.directories[i]
		entryPath := filepath.Join(root, entry.Name())
		if err := l.walkExtra(entryPath, book); err != nil {
			return err
		}
	}

	return nil
}

func (l *Type) convertToBook(in *bookMeta, out *db.Book) {
	// known volume ids
	out.Name = in.Name
	out.Writer = in.Writer
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

func (l *Type) convertToVolumes(volumeMetas []volumeMeta, known map[string]db.Volume) []db.Volume {
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
		}

		vol.Title = volMeta.Name
		vol.Volume = volMeta.ID
		vol.Files = convertArcFiles(volMeta.Files)
		vol.PageCount = len(vol.Files)

		vols[i] = vol
	}
	return vols
}
