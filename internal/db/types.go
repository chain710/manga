package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Library struct {
	ID       int64     `db:"id" json:"id"`
	CreateAt time.Time `db:"create_at" json:"create_at"`
	Name     string    `db:"name" json:"name"`
	Path     string    `db:"path" json:"path"`
}

type Book struct {
	ID        int64     `db:"id"`
	LibraryID int64     `db:"library_id"`
	CreateAt  time.Time `db:"create_at"`
	UpdateAt  time.Time `db:"update_at"`
	Path      string    `db:"path"`
	Name      string    `db:"name"`
	Writer    string    `db:"writer"`
	Volume    int       `db:"volume"`
	Summary   string    `db:"summary"`
	Files     BookFiles `db:"files"`
}

type BookFiles struct {
	Volumes []BookFile `json:"volumes"`
	Extras  []BookFile `json:"extras,omitempty"`
}

type BookFile struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Path string `json:"path"`
}

func (v BookFiles) Value() (driver.Value, error) {
	return json.Marshal(v)
}

func (v *BookFiles) Scan(src interface{}) error {
	switch val := src.(type) {
	case string:
		return json.Unmarshal([]byte(val), v)
	default:
		return errors.New("invalid src type")
	}
}
