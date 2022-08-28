package arc

import (
	internalstrings "github.com/chain710/manga/internal/strings"
	"github.com/klauspost/compress/zip"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestMy(t *testing.T) {
	zf, err := zip.OpenReader("testdata/1.zip")
	require.NoError(t, err)
	t.Logf("file count: %d", len(zf.File))
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expect      internalstrings.Set
		expectError bool
	}{
		{
			name: "zip",
			path: "testdata/1.zip",
			expect: internalstrings.NewSet(nil,
				filepath.Join("story1", "begin.txt"),
				filepath.Join("story2", "extra", "doc.txt"),
				"main.txt", "side.txt"),
		},
		{
			name: "rar",
			path: "testdata/2.rar",
			expect: internalstrings.NewSet(nil,
				filepath.Join("story1", "begin.txt"),
				filepath.Join("story2", "extra", "doc.txt"),
				"main.txt", "side.txt"),
		},
		{
			name:        "not exist",
			path:        "testdata/bad.zip",
			expectError: true,
		},
		{
			name:        "is dir",
			path:        "testdata/",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Open(tt.path)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			files := actual.GetFiles()
			actualPaths := internalstrings.NewSet(nil)
			for _, file := range files {
				actualPaths.Add(file.Name())
			}

			require.Equal(t, tt.expect.SortedList(), actualPaths.SortedList())
		})
	}
}

func TestRead(t *testing.T) {
	archive := "testdata/2.rar"
	a, err := Open(archive)
	require.NoError(t, err)
	gf, err := a.GetFile(filepath.Join("story2", "extra", "doc.txt"))
	require.NoError(t, err)
	data, err := a.ReadFileAt(gf.offset)
	require.NoError(t, err)
	require.Equal(t, "world", string(data))
}

func TestArchive_ReadFile(t *testing.T) {
	tests := []struct {
		name          string
		archiveFile   string
		fileInArchive string
		expectContent string
	}{
		{
			name:          "rar",
			archiveFile:   "testdata/2.rar",
			fileInArchive: "main.txt",
			expectContent: "hello",
		},
		{
			name:          "rar-1",
			archiveFile:   "testdata/2.rar",
			fileInArchive: filepath.Join("story2", "extra", "doc.txt"),
			expectContent: "world",
		},
		{
			name:          "zip",
			archiveFile:   "testdata/1.zip",
			fileInArchive: filepath.Join("story1", "begin.txt"),
			expectContent: "introduction",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := Open(tt.archiveFile)
			require.NoError(t, err)
			file, err := a.GetFile(tt.fileInArchive)
			require.NoError(t, err)
			actual, err := a.ReadFileAt(file.offset)
			require.NoError(t, err)
			require.Equal(t, tt.expectContent, string(actual))
		})
	}
}
