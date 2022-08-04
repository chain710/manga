package db

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
)

type Interface interface {
	// GetMigration for migrate
	GetMigration() (*migrate.Migrate, error)
	CreateLibrary(ctx context.Context, l *Library) error
	DeleteLibrary(ctx context.Context, opt DeleteLibraryOptions) error
	GetLibrary(ctx context.Context, id int64) (*Library, error)
	ListLibraries(ctx context.Context) ([]Library, error)
	PatchLibrary(ctx context.Context, opt PatchLibraryOptions) (*Library, error)
	GetBook(ctx context.Context, opt GetBookOptions) (*Book, error)
	CreateBook(ctx context.Context, b *Book) error
	UpdateBook(ctx context.Context, b *Book) error
	PatchBook(ctx context.Context, opt PatchBookOptions) (*Book, error)
	DeleteBook(ctx context.Context, options DeleteBookOptions) error
	ListBooks(ctx context.Context, opt ListBooksOptions) ([]Book, int, error)
	GetVolume(ctx context.Context, opt GetVolumeOptions) (*Volume, error)
	ListVolumes(ctx context.Context, opt ListVolumesOptions) ([]Volume, error)
	GetVolumeNeighbour(ctx context.Context, opt GetVolumeNeighbourOptions) (*int64, *int64, error)
	// TODO SetVolumeProgress remove
	SetVolumeProgress(ctx context.Context, opt VolumeProgressOptions) error
	BatchUpdateVolumeProgress(ctx context.Context, opt BatchUpdateVolumeProgressOptions) error
	SetVolumeThumbnail(ctx context.Context, thumbnail VolumeThumbnail) error
	GetVolumeThumbnail(ctx context.Context, opt GetVolumeThumbOptions) (*VolumeThumbnail, error)
	SetBookThumbnail(ctx context.Context, thumbnail BookThumbnail) error
	GetBookThumbnail(ctx context.Context, id int64) (*BookThumbnail, error)
}

const (
	ListBooksLeftJoinProgress  = "LeftJoinProgress"
	ListBooksRightJoinProgress = "RightJoinProgress"
	ListBookWithoutThumbnail   = "WithoutThumbnail"
	ListBooksOnly              = ""
)

type GetBookOptions struct {
	ID              int64
	Path            string
	WithoutProgress bool
}

type PatchBookOptions struct {
	ID      int64
	Name    string
	Writer  string
	Summary string
}

type ListBooksOptions struct {
	LibraryID *int64 // all lib if LibraryID == nil
	Offset    int
	Limit     int
	Sort      string
	Join      string
}

type GetVolumeOptions struct {
	ID int64
}

type GetVolumeNeighbourOptions struct {
	BookID int64
	Volume int
}

type VolumeProgressOptions struct {
	BookID   int64
	VolumeID int64
	Page     int
	Complete bool
}

const (
	UpdateVolumeProgressComplete = "Complete"
	UpdateVolumeProgressReset    = "Reset"
)

type BatchUpdateVolumeProgressOptions struct {
	IDs     []int64
	Operate string
}

type DeleteLibraryOptions struct {
	ID int64
}

type PatchLibraryOptions struct {
	ID     int64
	Name   string
	ScanAt Time
}

type ListVolumesOptions struct {
	WithoutThumbnail bool
}

type GetVolumeThumbOptions struct {
	ID     int64
	BookID *int64
}

type DeleteBookOptions struct {
	ID int64
}
