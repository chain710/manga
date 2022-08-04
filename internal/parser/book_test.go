package parser

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestParseBookName(t *testing.T) {
	tests := []struct {
		name        string
		bookString  string
		expect      *BookNameMeta
		expectError bool
	}{
		{
			name:       "normal",
			bookString: "[鐵腕女投手][高橋努][HK]",
			expect:     &BookNameMeta{Name: "鐵腕女投手", Writer: "高橋努"},
		},
		{
			name:       "no []",
			bookString: "Monster",
			expect:     &BookNameMeta{Name: "Monster"},
		},
		{
			name:        "empty",
			bookString:  "",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBookName(tt.bookString)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expect, got)
		})
	}
}

func TestParseBookVolumeBasic(t *testing.T) {
	file1 := filepath.Join("testdata2", "book - vol_2.zip")
	file1Info, err := os.Lstat(file1)
	require.NoError(t, err)
	tests := []struct {
		name   string
		path   string
		expect VolumeBasicMeta
	}{
		{
			name: "normal",
			path: file1,
			expect: VolumeBasicMeta{
				Name:    "book vol 2",
				Path:    file1,
				Size:    file1Info.Size(),
				ModTime: file1Info.ModTime(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseBookVolumeBasic(tt.path)
			require.NoError(t, err)
			require.Equal(t, actual, tt.expect)
		})
	}
}
