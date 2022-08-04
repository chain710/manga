package parser

import (
	"errors"
	"fmt"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//goland:noinspection RegExpRedundantEscape
var (
	bookNameRegex = regexp.MustCompile(`\[([^\]]+)\]`)
)

type BookNameMeta struct {
	Name   string
	Writer string
}

func (b *BookNameMeta) String() string {
	return fmt.Sprintf("%s By %s", b.Name, b.Writer)
}

func ParseBookName(name string) (*BookNameMeta, error) {
	name = strings.Trim(name, " \t\n")
	if len(name) == 0 {
		return nil, errors.New("invalid book name")
	}
	var meta BookNameMeta
	matches := bookNameRegex.FindAllStringSubmatch(name, -1)
	if len(matches) == 0 {
		meta.Name = name
		return &meta, nil
	}
	for i, match := range matches {
		if len(match) != 2 {
			return nil, fmt.Errorf("invalid match %v in name %s", match, name)
		}

		switch i {
		case 0:
			meta.Name = match[1]
		case 1:
			meta.Writer = match[1]
		default:
			continue
		}
	}

	return &meta, nil
}

type VolumeBasicMeta struct {
	Name    string // name or last digit
	ID      int    // Id as integer
	Path    string
	Size    int64
	ModTime time.Time
}

type VolumeMeta struct {
	VolumeBasicMeta
	Files []arc.File // files in archive
}

func ParseBookVolumeBasic(volumePath string) (VolumeBasicMeta, error) {
	//goland:noinspection RegExpRedundantEscape
	r := regexp.MustCompile(`[^\s\]\[_\-\.]+`)
	info, err := os.Lstat(volumePath)
	if err != nil {
		return VolumeBasicMeta{}, err
	}
	volName := stripExt(filepath.Base(volumePath))
	parts := r.FindAllString(volName, -1)
	return VolumeBasicMeta{Name: strings.Join(parts, " "), Path: volumePath, Size: info.Size(), ModTime: info.ModTime()}, nil
}

// ParseBookVolume parse single volume(archive file)
func ParseBookVolume(volumePath string, options *Options) (*VolumeMeta, error) {
	logger := log.With("vol", volumePath)
	basic, err := ParseBookVolumeBasic(volumePath)
	if err != nil {
		logger.Errorf("parse vol basic meta error: %s", err)
		return nil, err
	}
	archive, err := arc.Open(volumePath)
	if err != nil {
		logger.Errorf("open volume error %s", err)
		return nil, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer archive.Close()
	files := archive.GetFiles()
	filesInVol := make([]arc.File, 0, len(files))
	for _, file := range files {
		if options.isArcFileAllowed(file.Name()) {
			filesInVol = append(filesInVol, file)
		}
	}

	if options.SortArchiveFiles != nil {
		options.SortArchiveFiles(filesInVol)
	}
	return &VolumeMeta{
		VolumeBasicMeta: basic,
		Files:           filesInVol,
	}, nil
}

func stripExt(path string) string {
	ext := filepath.Ext(path)
	return path[:(len(path) - len(ext))]
}
