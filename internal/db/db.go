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
	// volume join
	VolumeCompactProgress      = "CompactVolumeProgress"
	VolumeLeftJoinBookProgress = "LeftJoinBookProgress"
	VolumeReading              = "VolumeReading"
	VolumeMustNotHaveThumb     = "VolumeMustNotHaveThumb"

	GetBookJoinVolumeProgress = "JoinVolumeProgress"
)

type GetBookOptions struct {
	ID   int64
	Path string
	Join string
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
	ID   int64
	Join string
}

type GetVolumeNeighbourOptions struct {
	BookID   int64
	Volume   int
	VolumeID int64
}

type VolumeProgressOptions struct {
	BookID   int64
	VolumeID int64
	Page     int
	Complete bool
}

const (
	UpdateVolumeProgressUpdate   = "Update"
	UpdateVolumeProgressComplete = "Complete"
	UpdateVolumeProgressReset    = "Reset"
)

type SetVolumeProgress struct {
	VolumeID int64
	Page     int
}

type BatchUpdateVolumeProgressOptions struct {
	SetVolumes []SetVolumeProgress
	Operate    string
}

func (b BatchUpdateVolumeProgressOptions) IDs() []int64 {
	ids := make([]int64, len(b.SetVolumes))
	for i := range b.SetVolumes {
		ids[i] = b.SetVolumes[i].VolumeID
	}
	return ids
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
	BookID    *int64
	LibraryID *int64
	Join      string
	Limit     int
}

type GetVolumeThumbOptions struct {
	ID     int64
	BookID *int64
}

type DeleteBookOptions struct {
	ID int64
}
