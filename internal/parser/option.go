package parser

import (
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/log"
	internalstrings "github.com/chain710/manga/internal/strings"
	"io/fs"
	"path/filepath"
	"strings"
)

type Options struct {
	AllowVolumeTypes   internalstrings.Set
	AllowHiddenVolume  bool
	AllowHiddenArcFile bool
	AllowArcFileTypes  internalstrings.Set
	SortArchiveFiles   func([]arc.File)
	SortVolumes        func([]VolumeMeta)
}

func (p *Options) isArcFileAllowed(path string) bool {
	base := filepath.Base(path)
	hidden := base[0:1] == "."
	if !p.AllowHiddenArcFile && hidden {
		return false
	}
	ext := filepath.Ext(path)
	return p.AllowArcFileTypes.Len() == 0 || p.AllowArcFileTypes.Contains(ext)
}

// classifyFiles return accept files and not accept files, hidden files are dropped
func (p *Options) classifyFiles(root string, files []fs.DirEntry) classifiedEntries {
	var ce classifiedEntries
	for i := range files {
		file := files[i]
		path := filepath.Join(root, file.Name())
		if !p.AllowHiddenVolume {
			if isHidden, err := isHiddenFile(path); err != nil || isHidden {
				if err != nil {
					log.Warnf("determine file %s hidden error: %s", path, err)
				}
				continue
			}
		}

		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (p.AllowVolumeTypes.Len() == 0 || p.AllowVolumeTypes.Contains(ext)) {
			ce.volumes = append(ce.volumes, file)
		} else if file.IsDir() {
			ce.directories = append(ce.directories, file)
		}
	}

	return ce
}

type Option func(*Options)

func DefaultOptions() Options {
	return Options{
		AllowVolumeTypes:  internalstrings.NewSet(strings.ToLower, ".zip", ".rar", ".7z"),
		AllowHiddenVolume: false,
		AllowArcFileTypes: internalstrings.NewSet(strings.ToLower, ".jpg", ".png", ".bmp"),
		SortArchiveFiles: func(files []arc.File) {
			SortSliceByDigit(files, extractArchiveFileDigits(files))
		},
		SortVolumes: func(metas []VolumeMeta) {
			SortSliceByDigit(metas, extractVolumeMetaDigits(metas))
		},
	}
}

func WithAllowVolumeTypes(types ...string) Option {
	return func(options *Options) {
		options.AllowVolumeTypes = internalstrings.NewSet(strings.ToLower, types...)
	}
}

func WithAllowArcFileTypes(types ...string) Option {
	return func(options *Options) {
		options.AllowArcFileTypes = internalstrings.NewSet(strings.ToLower, types...)
	}
}
