package db

import (
	"database/sql/driver"
	"errors"
	"time"
)

func NewTime(t time.Time) Time {
	copied, _ := time.Parse(time.RFC3339, t.Format(time.RFC3339))
	return Time{copied}
}

// Time store time in seconds
type Time struct {
	t time.Time
}

var _ driver.Valuer = Time{}

func (t Time) Value() (driver.Value, error) {
	return t.t, nil
}

func (t *Time) Scan(src interface{}) error {
	switch val := src.(type) {
	case time.Time:
		t.t = val
		return nil
	default:
		return errors.New("invalid src type")
	}
}

func (t *Time) EqualTime(other time.Time) bool {
	t2 := NewTime(other)
	return t.t.Equal(t2.t)
}

func (t *Time) IsZero() bool {
	return t.t.IsZero()
}
