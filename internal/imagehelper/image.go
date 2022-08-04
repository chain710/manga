package imagehelper

import (
	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	"image"
	"io"
)

func DecodeConfig(r io.Reader) (image.Config, string, error) {
	return image.DecodeConfig(r)
}

// GetChroma return chroma of image [0, 1]
func GetChroma(img image.Image, sampleHeight int) float64 {
	originalSize := img.Bounds().Size()
	height := originalSize.Y
	if height > sampleHeight {
		height = sampleHeight
	}
	resized := imaging.Resize(img, 0, height, imaging.Lanczos)
	resizeBounds := resized.Bounds()
	pixels := resizeBounds.Dx() * resizeBounds.Dy()
	var chromaSum float64
	for y := resizeBounds.Min.Y; y < resizeBounds.Max.Y; y++ {
		for x := resizeBounds.Min.X; x < resizeBounds.Max.X; x++ {
			i, _ := colorful.MakeColor(resized.At(x, y))
			_, c, _ := i.Hcl()
			chromaSum += c
		}
	}

	return chromaSum / float64(pixels)
}
