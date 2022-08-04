package arc

import (
	"github.com/gen2brain/go-unarr"
	"io/fs"
)

type realArchive struct {
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

// ReadFile return whole content of file
func (f *realArchive) ReadFile(file File) ([]byte, error) {
	err := f.impl.EntryAt(file.offset)
	if err != nil {
		return nil, err
	}

	return f.impl.ReadAll()
}

func (f *realArchive) ReadFileAt(offset int64) ([]byte, error) {
	err := f.impl.EntryAt(offset)
	if err != nil {
		return nil, err
	}

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
	if f.impl == nil {
		return nil
	}
	err := f.impl.Close()
	f.impl = nil
	return err
}
