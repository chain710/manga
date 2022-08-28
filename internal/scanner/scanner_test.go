package scanner

import (
	"context"
	"github.com/chain710/manga/internal/db"
	dbmocks "github.com/chain710/manga/internal/db/mocks"
	"github.com/chain710/workqueue"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
	"time"
)

type testBook struct {
	op string
	bookNameMeta
	id      int64
	lib     int64
	path    string
	volumes []testVolume
	extras  []testVolume
}

// testVolume for test compare; ignore fields like offset,modtime
type testVolume struct {
	volume int // Id as integer
	path   string
	files  []string
}

func newTestBook(item *BookItem) testBook {
	var b testBook
	b.op = item.Op
	b.id = item.Book.ID
	b.lib = item.Book.LibraryID
	b.path = item.Book.Path
	b.Name = item.Book.Name
	b.Writer = item.Book.Writer

	f := func(volumes []db.Volume) []testVolume {
		var ret []testVolume
		for _, vol := range volumes {
			var files []string
			for _, f := range vol.Files {
				files = append(files, f.Path)
			}
			ret = append(ret, testVolume{
				volume: vol.Volume,
				path:   vol.Path,
				files:  files,
			})
		}
		return ret
	}

	b.volumes = f(item.Book.Volumes)
	b.extras = f(item.Book.Extras)
	return b
}

func TestType_Scan(t *testing.T) {
	root := filepath.Join("testdata", "scan", "lib1")
	now := time.Now()
	clk = workqueue.NewFakeClock(now)
	libs := []db.Library{
		{ID: 1, Path: root},
	}
	books := []db.Book{
		{BookTable: db.BookTable{ID: 1, Path: "fake1"}}, // should be deleted
	}
	mdb := dbmocks.NewInterface(t)
	mdb.On("ListBooks", mock.Anything, mock.Anything).
		Once().Return(books, len(books), nil)
	mdb.On("ListLibraries", mock.Anything).
		Once().
		Return(libs, nil)
	mdb.On("PatchLibrary", mock.Anything, mock.Anything).
		Once().Return(nil, nil)
	// not found any book
	mdb.On("GetBook", mock.Anything, mock.Anything).
		Return(nil, nil)
	q := workqueue.NewRetryQueue("test", clk)
	scanner := New(q, mdb)
	require.NoError(t, scanner.Scan(context.TODO()))
	q.ShutDown()

	expectBookPath := filepath.Join(root, "[book1][writer]")
	expectBook := map[string]testBook{
		"fake1": {op: OpDelete, id: 1, path: "fake1"},
		expectBookPath: {
			bookNameMeta: bookNameMeta{Name: "book1", Writer: "writer"},
			op:           OpNew,
			id:           0,
			lib:          1,
			path:         expectBookPath,
			volumes: []testVolume{
				{volume: 1, path: filepath.Join(expectBookPath, "2.zip"), files: []string{"cat.jpg"}},
				{volume: 2, path: filepath.Join(expectBookPath, "3.zip"), files: []string{"cat.jpg"}},
			},
			extras: []testVolume{
				{volume: 0, path: filepath.Join(expectBookPath, "extra1", "album-1.zip"), files: []string{"cat.jpg"}},
				{volume: 0, path: filepath.Join(expectBookPath, "extra2", "album-2.zip"), files: []string{"cat.jpg"}},
			},
		},
	}
	actualBook := make(map[string]testBook)
	for {
		item, shutdown := q.Get()
		if shutdown {
			break
		}

		bi := item.(*BookItem)
		actualBook[bi.Book.Path] = newTestBook(bi)
		q.Done(item, nil)
	}

	require.Equal(t, expectBook, actualBook)
}
