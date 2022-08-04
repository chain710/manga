package scanner

import (
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"time"
)

const (
	OpDelete = "delete"
	OpNew    = "new"
	OpUpdate = "update"
)

type BookItem struct {
	Book db.Book
	Op   string
}

func (b *BookItem) IsReplaceable() bool {
	return true
}

func (b *BookItem) Index() interface{} {
	return b.Book.Path
}

type bookNameMeta struct {
	Name   string
	Writer string
}

type volumeBasicMeta struct {
	Name    string // name or last digit
	ID      int    // Id as integer
	Path    string
	Size    int64
	ModTime time.Time
}

type volumeMeta struct {
	volumeBasicMeta
	Files []arc.File // files in archive
}

type bookMeta struct {
	bookNameMeta
	Volumes []volumeMeta
	Path    string
	ModTime time.Time
	Extras  []volumeMeta
}
