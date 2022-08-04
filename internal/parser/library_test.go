package parser

import (
	internalstrings "github.com/chain710/manga/internal/strings"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// testVolume for test compare; ignore fields like offset,modtime
type testVolume struct {
	basic VolumeBasicMeta
	files []string
}

func newTestVolume(name string, id int, path string, files ...string) testVolume {
	info, err := os.Lstat(path)
	if err != nil {
		panic(err)
	}
	sort.Strings(files)
	return testVolume{
		basic: VolumeBasicMeta{
			Name:    name,
			ID:      id,
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		},
		files: files,
	}
}

func newTestVolumeFromMeta(meta VolumeMeta) testVolume {
	ss := internalstrings.NewSet(nil)
	for _, f := range meta.Files {
		ss.Add(f.Name())
	}
	return testVolume{
		basic: meta.VolumeBasicMeta,
		files: ss.SortedList(),
	}
}

type testBook struct {
	vols   []testVolume
	extras []testVolume
}

func TestWalkLibrary(t *testing.T) {
	root := "testdata"
	actualBooks := make(map[string]testBook)
	factory := LibraryWalker{
		Predict: func(*BookMeta) bool { return true },
		Handle: func(b *BookMeta) {
			tb := testBook{
				vols:   []testVolume{},
				extras: []testVolume{},
			}
			for _, vol := range b.Volumes {
				tb.vols = append(tb.vols, newTestVolumeFromMeta(vol))
			}
			for _, vol := range b.Extras {
				tb.extras = append(tb.extras, newTestVolumeFromMeta(vol))
			}

			actualBooks[b.Name.Name] = tb
		},
	}

	err := WalkLibrary(root, func() LibraryWalker {
		return factory
	})
	require.NoError(t, err)
	expect := map[string]testBook{
		"book1": {
			vols: []testVolume{
				newTestVolume("2", 1, filepath.Join(root, "[book1][writer]", "2.zip"), "cat.jpg"),
				newTestVolume("3", 2, filepath.Join(root, "[book1][writer]", "3.zip"), "cat.jpg"),
			},
			extras: []testVolume{
				newTestVolume("album 1", 0, filepath.Join(root, "[book1][writer]", "extra1", "album-1.zip"), "cat.jpg"),
				newTestVolume("album 2", 0, filepath.Join(root, "[book1][writer]", "extra2", "album-2.zip"), "cat.jpg"),
			},
		},
	}
	require.Equal(t, expect, actualBooks)
}
