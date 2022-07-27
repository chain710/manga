package parser

import (
	"errors"
	"fmt"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/log"
	internalstrings "github.com/chain710/manga/internal/strings"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

type VolumeOptions struct {
	AcceptFileTypes internalstrings.Set
	SortFiles       func([]arc.File)
}

type VolumeOption func(*VolumeOptions)

func VolumeOptionsDefault(opt *VolumeOptions) {
	opt.AcceptFileTypes = internalstrings.NewSet(".jpg", ".bmp", ".png")
	opt.SortFiles = SortByName
}

type BookVolumeBasicMeta struct {
	Name string // name or last digit
	ID   int    // Id as integer
	Path string
}

type BookVolumeMeta struct {
	BookVolumeBasicMeta
	Files []arc.File
}

func ParseBookVolumeBasic(volumePath string) BookVolumeBasicMeta {
	volName := filepath.Base(volumePath)
	digit := regexp.MustCompile(`\d+`)
	id := stripExt(volName)
	digits := digit.FindAllString(id, -1)
	var idInt int
	if len(digits) != 0 {
		id = digits[len(digits)-1]
		// ignore this error
		idInt, _ = strconv.Atoi(id)
	}

	return BookVolumeBasicMeta{Name: id, ID: idInt, Path: volumePath}
}

// ParseBookVolume parse single volume(archive file)
func ParseBookVolume(volumePath string, options ...VolumeOption) (*BookVolumeMeta, error) {
	logger := log.With("vol", volumePath)
	var opt VolumeOptions
	for _, apply := range options {
		apply(&opt)
	}

	basic := ParseBookVolumeBasic(volumePath)
	archive, err := arc.Open(volumePath)
	if err != nil {
		logger.Errorf("open volume error %s", err)
		return nil, err
	}

	files := archive.GetFiles()
	filesInVol := make([]arc.File, 0, len(files))
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if opt.AcceptFileTypes.Len() == 0 || opt.AcceptFileTypes.Contains(ext) {
			filesInVol = append(filesInVol, file)
		}
	}

	if opt.SortFiles != nil {
		opt.SortFiles(filesInVol)
	}
	return &BookVolumeMeta{
		BookVolumeBasicMeta: basic,
		Files:               filesInVol,
	}, nil
}

func stripExt(path string) string {
	ext := filepath.Ext(path)
	return path[:(len(path) - len(ext))]
}
