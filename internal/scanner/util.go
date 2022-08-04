package scanner

import (
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"os"
	"path/filepath"
	"strings"
)

func stripExt(path string) string {
	ext := filepath.Ext(path)
	return path[:(len(path) - len(ext))]
}

func parseBookName(in string) bookNameMeta {
	parts := regexDelimiter.FindAllString(in, -1)
	var bn bookNameMeta
	if len(parts) > 0 {
		bn.Name = parts[0]
	}
	if len(parts) > 1 {
		bn.Writer = parts[1]
	}

	return bn
}

func parseVolumeBasic(path string) (volumeBasicMeta, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return volumeBasicMeta{}, err
	}
	volName := stripExt(filepath.Base(path))
	parts := regexDelimiter.FindAllString(volName, -1)
	return volumeBasicMeta{Name: strings.Join(parts, " "), Path: path, Size: info.Size(), ModTime: info.ModTime()}, nil
}

func convertArcFiles(files []arc.File) []db.VolumeFile {
	vf := make([]db.VolumeFile, len(files))
	for i, f := range files {
		vf[i] = db.VolumeFile{
			Path:   f.Name(),
			Offset: f.Offset(),
			Size:   f.Size(),
		}
	}
	return vf
}
