package arc

import (
	"github.com/chain710/gounrar"
	"github.com/chain710/manga/internal/log"
	"io"
	"io/fs"
	"sync"
)

func newRARArchive(path string, skipReadingFiles bool) (*rarArchive, error) {
	a, err := gounrar.Open(path)
	if err != nil {
		return nil, err
	}

	log.Debugf("open archive: %s", path)
	var files []File
	if !skipReadingFiles {
		for {
			hdr, err := a.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Errorf("reading rar %s files error: %s", path, err)
				return nil, err
			}

			if hdr.IsDir() {
				continue
			}

			files = append(files, File{
				name:    hdr.FileName,
				offset:  hdr.BlockPos,
				modTime: hdr.ModTime,
				size:    int(hdr.UnpackSize),
			})
		}
	}
	rar := &rarArchive{
		path:    path,
		archive: a,
		files:   files,
	}
	return rar, nil
}

type rarArchive struct {
	mu      sync.Mutex // protect archive
	archive *gounrar.Archive
	files   []File // constant, no lock
	path    string
}

func (u *rarArchive) Path() string {
	return u.path
}

func (u *rarArchive) GetFiles() []File {
	return u.files
}

func (u *rarArchive) ReadFileAt(offset int64) ([]byte, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, err := u.archive.SeekPos(offset)
	if err != nil {
		log.Errorf("read rar file at %d error: %s", offset, err)
		return nil, err
	}

	return u.archive.ReadAll()
}

func (u *rarArchive) GetFile(path string) (*File, error) {
	for i := range u.files {
		if u.files[i].Name() == path {
			return &u.files[i], nil
		}
	}
	return nil, fs.ErrNotExist
}

func (u *rarArchive) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	var err error
	if u.archive != nil {
		err = u.archive.Close()
	}
	u.archive = nil
	return err
}
