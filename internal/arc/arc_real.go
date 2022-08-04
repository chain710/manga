package arc

import (
	"github.com/chain710/manga/internal/log"
	"github.com/gen2brain/go-unarr"
	"io/fs"
	"sync"
)

type realArchive struct {
	mu    sync.Mutex // protect impl
	path  string
	impl  *unarr.Archive
	files []File
}

var _ Archive = &realArchive{}

func (f *realArchive) Path() string {
	return f.path
}

func (f *realArchive) GetFiles() []File {
	return f.files
}

func (f *realArchive) ReadFileAt(offset int64) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	err := f.impl.EntryAt(offset)
	if err != nil {
		log.Debugf("read file %s at %d error: %s", f.path, offset, err)
		return nil, err
	}

	log.Debugf("read file %s at %d ok", f.path, offset)
	return f.impl.ReadAll()
}

func (f *realArchive) GetFile(path string) (*File, error) {
	// TODO use index to speedup
	for i := range f.files {
		if f.files[i].Name() == path {
			return &f.files[i], nil
		}
	}
	return nil, fs.ErrNotExist
}

func (f *realArchive) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.impl == nil {
		return nil
	}

	log.Debugf("close archive: %s", f.path)
	err := f.impl.Close()
	f.impl = nil
	return err
}
