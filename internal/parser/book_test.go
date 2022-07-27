package parser

import (
	"github.com/stretchr/testify/require"
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
	tests := []struct {
		name   string
		path   string
		expect BookVolumeBasicMeta
	}{
		{
			name: "normal",
			path: filepath.Join("dir", "my vol 01.zip03"),
			expect: BookVolumeBasicMeta{
				Name: "01",
				ID:   1,
				Path: filepath.Join("dir", "my vol 01.zip03"),
			},
		},
		{
			name: "without digit part",
			path: filepath.Join("dir", "mybook.zip"),
			expect: BookVolumeBasicMeta{
				Name: "mybook",
				ID:   0,
				Path: filepath.Join("dir", "mybook.zip"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseBookVolumeBasic(tt.path)
			require.Equal(t, actual, tt.expect)
		})
	}
}
