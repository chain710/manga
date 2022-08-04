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
	Name     string `db:"name" json:"name"`
	Path     string `db:"path" json:"path"`
}

type Book struct {
	ID        int64  `db:"id"`
	LibraryID int64  `db:"library_id"`
	CreateAt  Time   `db:"create_at"`
	UpdateAt  Time   `db:"update_at"`
	PathModAt Time   `db:"path_mod_at"`
	Path      string `db:"path"`
	Name      string `db:"name"`
	Writer    string `db:"writer"`
	Volume    int    `db:"volume"`
	Summary   string `db:"summary"`
	Volumes   []Volume
	Extras    []Volume
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
	ID       int64          `db:"id"`
	BookID   int64          `db:"book_id"`
	CreateAt Time           `db:"create_at"`
	Path     string         `db:"path"`
	Title    string         `db:"title"`
	Cover    string         `db:"cover"`
	Volume   int            `db:"volume"` // 0 = extra
	Files    VolumeFileList `db:"files"`
}

func (v *Volume) DeepCopy() *Volume {
	n := *v
	n.Files = v.Files[:]
	return &n
}

type VolumeFile struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
}

type VolumeFileList []VolumeFile

var _ driver.Valuer = VolumeFileList{}

//goland:noinspection GoMixedReceiverTypes
func (v VolumeFileList) Value() (driver.Value, error) {
	return json.Marshal(v)
}

//goland:noinspection GoMixedReceiverTypes
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

type VolumeProgress struct {
	CreateAt Time  `db:"create_at"`
	UpdateAt Time  `db:"update_at"`
	BookID   int64 `db:"book_id"`
	VolumeID int64 `db:"volume_id"`
	Complete bool  `db:"complete"`
	Page     int   `db:"page"`
}
