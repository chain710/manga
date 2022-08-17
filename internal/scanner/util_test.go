package scanner

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func Test_parseBookName(t *testing.T) {
	tests := []struct {
		name       string
		bookString string
		expect     bookNameMeta
	}{
		{
			name:       "normal",
			bookString: "[鐵腕女投手][高橋努][HK]",
			expect:     bookNameMeta{Name: "鐵腕女投手", Writer: "高橋努"},
		},
		{
			name:       "normal 2",
			bookString: "Monster - puze",
			expect:     bookNameMeta{Name: "Monster", Writer: "puze"},
		},
		{
			name:       "empty",
			bookString: "",
			expect:     bookNameMeta{Name: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBookName(tt.bookString)
			require.Equal(t, tt.expect, got)
		})
	}
}

func Test_parseVolumeBasic(t *testing.T) {
	file1 := filepath.Join("testdata", "parsevolbasic", "book - vol_2.zip")
	file1Info, err := os.Lstat(file1)
	require.NoError(t, err)
	tests := []struct {
		name        string
		path        string
		expect      volumeBasicMeta
		expectError bool
	}{
		{
			name: "normal",
			path: file1,
			expect: volumeBasicMeta{
				Name:    "book vol 2",
				Path:    file1,
				Size:    file1Info.Size(),
				ModTime: file1Info.ModTime(),
			},
		},
		{
			name:        "error",
			path:        "path-not-exist",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseVolumeBasic(tt.path)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, actual, tt.expect)
			}
		})
	}
}
