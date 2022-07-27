package parser

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"sort"
	"testing"
)

func TestWalkLibrary(t *testing.T) {
	root := "testdata"
	bookMetas := make(map[string]*BookMeta)
	err := WalkLibrary(root,
		func(b *BookMeta) {
			sort.Slice(b.Volumes, func(i, j int) bool {
				return b.Volumes[i].Name < b.Volumes[j].Name
			})
			sort.Slice(b.Extras, func(i, j int) bool {
				return b.Extras[i].Name < b.Extras[j].Name
			})
			bookMetas[b.Name.Name] = b
		},
		LibraryOptionAcceptFileTypes(".zip"),
	)
	require.NoError(t, err)
	expect := map[string]*BookMeta{
		"book1": {
			Name: BookNameMeta{Name: "book1"},
			Volumes: []BookVolumeBasicMeta{
				{Name: "1", ID: 1, Path: filepath.Join(root, "book1", "1.zip")},
				{Name: "2", ID: 2, Path: filepath.Join(root, "book1", "2.zip")},
			},
			Path:   filepath.Join(root, "book1"),
			Extras: []BookVolumeBasicMeta{},
		},
		"book2": {
			Name: BookNameMeta{Name: "book2", Writer: "b"},
			Volumes: []BookVolumeBasicMeta{
				{Name: "a", ID: 0, Path: filepath.Join(root, "[book2][b]", "a.zip")},
			},
			Path: filepath.Join(root, "[book2][b]"),
			Extras: []BookVolumeBasicMeta{
				{Name: "album", ID: 0, Path: filepath.Join(root, "[book2][b]", "extra1", "album.zip")},
				{Name: "new", ID: 0, Path: filepath.Join(root, "[book2][b]", "extra2", "new.zip")},
			},
		},
	}
	require.Equal(t, expect, bookMetas)
}
