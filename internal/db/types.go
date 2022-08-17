package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"os"
)

type Library struct {
	ID       int64  `db:"id" json:"id"`
	CreateAt Time   `db:"create_at" json:"create_at"`
	ScanAt   Time   `db:"scan_at" json:"scan_at"`
	Name     string `db:"name" json:"name"`
	Path     string `db:"path" json:"path"`
}

type Book struct {
	BookTable
	Volumes  []Volume      `json:"volumes,omitempty"`
	Extras   []Volume      `json:"extras,omitempty"`
	Progress *BookProgress `json:"progress,omitempty"`
}

func (b *Book) GetVolume(id int64) (*Volume, error) {
	return findInVolumes(id, b.Volumes)
}

func (b *Book) GetExtra(id int64) (*Volume, error) {
	return findInVolumes(id, b.Extras)
}

func (b *Book) GetVolumeIDs() []int64 {
	ids := make([]int64, 0, len(b.Volumes)+len(b.Extras))
	f := func(vols []Volume) {
		for _, vol := range vols {
			if vol.ID >= 0 {
				ids = append(ids, vol.ID)
			}
		}
	}

	f(b.Volumes)
	f(b.Extras)
	return ids
}

func (b *Book) SyncBookID() {
	for i := range b.Volumes {
		b.Volumes[i].BookID = b.ID
	}
	for i := range b.Extras {
		b.Extras[i].BookID = b.ID
	}
}

type Volume struct {
	VolumeTable
	BookName string          `json:"book_name,omitempty"`
	Writer   string          `json:"writer,omitempty"`
	Progress *VolumeProgress `json:"progress,omitempty"`
}

func (v *Volume) DeepCopy() *Volume {
	n := *v
	n.Files = v.Files[:]

	n.Progress = v.Progress.DeepCopy()
	return &n
}

type VolumeFile struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
}

type VolumeFileList []VolumeFile

var _ driver.Valuer = VolumeFileList{}

func (v VolumeFileList) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *VolumeFileList) Scan(src interface{}) error {
	switch val := src.(type) {
	case string:
		return json.Unmarshal([]byte(val), v)
	default:
		return errors.New("invalid src type")
	}
}

func findInVolumes(id int64, files []Volume) (*Volume, error) {
	for i := range files {
		file := files[i]
		if id == file.ID {
			return &file, nil
		}
	}

	return nil, os.ErrNotExist
}

type BookProgress struct {
	BookID   int64 `db:"book_id" json:"book_id"`
	UpdateAt Time  `db:"update_at" json:"update_at"`
	// Volume now reading volume
	Volume int `db:"volume" json:"volume"`
	// VolumeID now reading volume id
	VolumeID  int64  `db:"volume_id" json:"volume_id"`
	Title     string `db:"title" json:"title"`
	Page      int    `db:"page" json:"page"`
	PageCount int    `db:"page_count" json:"page_count"`
}

type VolumeThumbnail struct {
	ID        int64  `db:"id"`
	Hash      string `db:"hash"`
	Thumbnail []byte `db:"thumbnail"`
}

type BookThumbnail struct {
	ID        int64  `db:"id"`
	Hash      string `db:"hash"`
	Thumbnail []byte `db:"thumbnail"`
}

type VolumeProgress struct {
	CreateAt Time `json:"create_at"`
	UpdateAt Time `json:"update_at"`
	Page     int  `json:"page"`
}

func (vp *VolumeProgress) DeepCopy() *VolumeProgress {
	return &VolumeProgress{
		CreateAt: vp.CreateAt,
		UpdateAt: vp.UpdateAt,
		Page:     vp.Page,
	}
}
