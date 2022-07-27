package arc

import (
	internalstrings "github.com/chain710/manga/internal/strings"
	"github.com/stretchr/testify/require"
	"testing"
)

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
			expect: internalstrings.NewSet("story1/begin.txt",
				"story2/extra/doc.txt",
				"main.txt", "side.txt"),
		},
		{
			name: "rar",
			path: "testdata/2.rar",
			expect: internalstrings.NewSet("story1/begin.txt",
				"story2/extra/doc.txt",
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
			actualPaths := internalstrings.NewSet()
			for _, file := range files {
				actualPaths.Add(file.Name())
			}

			require.Equal(t, tt.expect, actualPaths)
		})
	}
}

func TestRead(t *testing.T) {
	archive := "testdata/2.rar"
	a, err := Open(archive)
	require.NoError(t, err)
	gf, err := a.GetFile("story2/extra/doc.txt")
	require.NoError(t, err)
	data, err := a.ReadFile(*gf)
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
			fileInArchive: "story2/extra/doc.txt",
			expectContent: "world",
		},
		{
			name:          "zip",
			archiveFile:   "testdata/1.zip",
			fileInArchive: "story1/begin.txt",
			expectContent: "introduction",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := Open(tt.archiveFile)
			require.NoError(t, err)
			file, err := a.GetFile(tt.fileInArchive)
			require.NoError(t, err)
			actual, err := a.ReadFile(*file)
			require.NoError(t, err)
			require.Equal(t, tt.expectContent, string(actual))
		})
	}
}
