package imagehelper

import (
	"bytes"
	"errors"
	"github.com/chain710/manga/internal/arc"
	"github.com/chain710/manga/internal/db"
	"github.com/chain710/manga/internal/log"
	"github.com/disintegration/imaging"
	"image"
	"sort"
)

type VolumeThumbnailOptions struct {
	SampleHeight   int
	HeadCandidates int
	TailCandidates int
}

type imageCandidate struct {
	path   string
	image  image.Image
	chroma float64
	index  int // index in slice
}

const (
	minChroma = 0.01
)

// GetVolumeCover return most likely thumbnail(book cover) image in files
func GetVolumeCover(archive arc.Archive, files []db.VolumeFile, opt VolumeThumbnailOptions) (image.Image, error) {
	fileCount := len(files)
	if fileCount == 0 {
		return nil, errors.New("no files found")
	}

	total := opt.HeadCandidates + opt.TailCandidates
	candidates := make([]db.VolumeFile, 0, total)
	if fileCount <= total {
		candidates = files
	} else {
		candidates = append(candidates, files[0:opt.HeadCandidates]...)
		candidates = append(candidates, files[fileCount-opt.TailCandidates:fileCount]...)
	}

	var imageCandidates []imageCandidate
	logger := log.With("archive", archive.Path())
	for i, file := range candidates {
		data, err := archive.ReadFileAt(file.Offset)
		if err != nil {
			logger.Errorf("read file %s in archive error: %s", file.Path, err)
			continue
		}

		img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
		if err != nil {
			logger.Errorf("decode image %s error: %s", file.Path, err)
			continue
		}

		chroma := GetChroma(img, opt.SampleHeight)
		if chroma < minChroma {
			chroma = minChroma
		}
		imageCandidates = append(imageCandidates, imageCandidate{
			path:   file.Path,
			image:  img,
			chroma: chroma,
			index:  i,
		})
	}

	if len(imageCandidates) == 0 {
		return nil, errors.New("no candidates found")
	}

	sort.Slice(imageCandidates, func(i, j int) bool {
		// reverse
		if imageCandidates[i].chroma != imageCandidates[j].chroma {
			return imageCandidates[i].chroma > imageCandidates[j].chroma
		}

		return imageCandidates[i].index > imageCandidates[j].index
	})
	return imageCandidates[0].image, nil
}
