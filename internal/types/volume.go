package types

import (
	"errors"
	"github.com/chain710/manga/internal/db"
)

type Volume struct {
	db.Volume
	NextVolumeID *int64 `json:"next_volume_id,omitempty"`
	PrevVolumeID *int64 `json:"prev_volume_id,omitempty"`
}

type VolumeProgress struct {
	ID   int64 `json:"id"`
	Page int   `json:"page"`
}

type UpdateVolumeProgressRequest struct {
	Volumes  []VolumeProgress `json:"volumes"`
	Operator string           `json:"op"`
}

func (r *UpdateVolumeProgressRequest) Validate() error {
	if len(r.Volumes) == 0 {
		return errors.New("empty ids")
	}

	if r.Operator != db.UpdateVolumeProgressReset &&
		r.Operator != db.UpdateVolumeProgressComplete &&
		r.Operator != db.UpdateVolumeProgressUpdate {
		return errors.New("invalid operator")
	}

	return nil
}
