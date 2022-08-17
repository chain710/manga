package serve

import (
	"errors"
	"github.com/hashicorp/go-multierror"
)

type pageRect struct {
	Page   int     `json:"page"`
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Validate TODO use go-playground/validator
func (r *pageRect) Validate() error {
	var merr *multierror.Error
	if r.Page < 1 {
		merr = multierror.Append(merr, errors.New("page < 1"))
	}
	if r.Top < 0 {
		merr = multierror.Append(merr, errors.New("top < 0"))
	}
	if r.Left < 0 {
		merr = multierror.Append(merr, errors.New("left < 0"))
	}

	if r.Width <= 0 {
		merr = multierror.Append(merr, errors.New("width <= 0"))
	}

	if r.Height <= 0 {
		merr = multierror.Append(merr, errors.New("height <= 0"))
	}

	if r.Width+r.Left > 1 {
		merr = multierror.Append(merr, errors.New("exceed width limit"))
	}

	if r.Height+r.Top > 1 {
		merr = multierror.Append(merr, errors.New("exceed height limit"))
	}

	return merr.ErrorOrNil()
}
