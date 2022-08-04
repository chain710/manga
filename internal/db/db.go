package db

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
)

type GetBookOptions struct {
	ID   int64
	Path string
}

type ListBooksOptions struct {
	LibraryID int64
}

type GetVolumeOptions struct {
	ID int64
}

type VolumeProgressOptions struct {
	BookID   int64
	VolumeID int64
	Page     int
	Complete bool
}

type Interface interface {
	// GetMigration for migrate
	GetMigration() (*migrate.Migrate, error)
	CreateLibrary(ctx context.Context, l *Library) error
	//DeleteLibrary(ctx context.Context,l Library) error
	GetLibrary(ctx context.Context, id int) (*Library, error)
	ListLibraries(ctx context.Context) ([]Library, error)
	GetBook(ctx context.Context, opt GetBookOptions) (*Book, error)
	CreateBook(ctx context.Context, b *Book) error
	UpdateBook(ctx context.Context, b *Book) error
	//DeleteBook(ctx context.Context,b Book) error
	ListBooks(ctx context.Context, opt *ListBooksOptions) ([]Book, error)
	GetVolume(ctx context.Context, opt GetVolumeOptions) (*Volume, error)
	SetVolumeProgress(ctx context.Context, opt VolumeProgressOptions) error
}
