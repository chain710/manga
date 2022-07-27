package arc

import (
	"github.com/gen2brain/go-unarr"
	"io"
	"io/fs"
	"strings"
)

//type OpenOption func(options *OpenOptions)
//type Sort func([]File) sort.Interface
//
//type OpenOptions struct {
//	AcceptFileTypes internalstrings.Set
//	Sort            Sort
//}
//
//func OpenOptionDefault(options *OpenOptions) {
//	options.AcceptFileTypes = internalstrings.NewSet("jpg", "bmp", "png")
//	options.Sort = SortByName
//}

func Open(path string) (*Archive, error) {
	a, err := unarr.NewArchive(path)
	if err != nil {
		return nil, err
	}

	var files []File
	for {
		entryErr := a.Entry()
		if entryErr != nil {
			if entryErr == io.EOF {
				break
			}
			return nil, entryErr
		}

		files = append(files, File{
			name:    getName(a.Name(), a.RawName()),
			offset:  a.Offset(),
			modTime: a.ModTime(),
			size:    a.Size(),
		})
	}

	return &Archive{
		impl:  a,
		files: files,
	}, nil
}

type Archive struct {
	impl  *unarr.Archive
	files []File
}

func (f *Archive) GetFiles() []File {
	return f.files
}

// ReadFile return whole content of file
func (f *Archive) ReadFile(file File) ([]byte, error) {
	err := f.impl.EntryAt(file.offset)
	if err != nil {
		return nil, err
	}

	return f.impl.ReadAll()
}

func (f *Archive) GetFile(path string) (*File, error) {
	for i := range f.files {
		if f.files[i].Name() == path {
			return &f.files[i], nil
		}
	}
	return nil, fs.ErrNotExist
}

func (f *Archive) Close() error {
	return f.impl.Close()
}

func getName(name string, raw string) string {
	name = strings.ReplaceAll(name, `\`, `/`)
	raw = strings.ReplaceAll(raw, `\`, `/`)
	if raw != "" {
		return raw
	}

	return name
}
