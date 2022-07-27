package arc

import (
	"time"
)

// File represent file in archive
type File struct {
	name    string // name is file path in archive
	offset  int64
	modTime time.Time
	size    int // size of uncompressed data
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Offset() int64 {
	return f.offset
}

func (f *File) Size() int64 {
	return int64(f.size)
}

func (f *File) ModTime() time.Time {
	return f.modTime
}
