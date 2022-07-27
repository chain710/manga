package db

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
)

type ListBooksOptions struct {
	LibraryID int64
}

type Interface interface {
	// GetMigration for migrate
	GetMigration() (*migrate.Migrate, error)
	CreateLibrary(ctx context.Context, l *Library) error
	//DeleteLibrary(ctx context.Context,l Library) error
	GetLibrary(ctx context.Context, id int) (*Library, error)
	ListLibraries(ctx context.Context) ([]Library, error)
	CreateBook(ctx context.Context, b *Book) error
	UpdateBook(ctx context.Context, b *Book) error
	//DeleteBook(ctx context.Context,b Book) error
	ListBooks(ctx context.Context, opt *ListBooksOptions) ([]Book, error)
}
