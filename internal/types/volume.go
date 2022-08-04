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

type UpdateVolumeProgressRequest struct {
	VolumeIDs []int64 `json:"volume_ids"`
	Operator  string  `json:"op"`
}

func (r *UpdateVolumeProgressRequest) Validate() error {
	if len(r.VolumeIDs) == 0 {
		return errors.New("empty ids")
	}

	if r.Operator != db.UpdateVolumeProgressReset && r.Operator != db.UpdateVolumeProgressComplete {
		return errors.New("invalid operator")
	}

	return nil
}
