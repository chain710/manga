package types

import (
	"errors"
	"os"
)

type AddLibraryRequest struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (a *AddLibraryRequest) Validate() error {
	if a.Name == "" {
		return errors.New("empty name")
	}

	info, err := os.Stat(a.Path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("path should be directory")
	}

	return nil
}
