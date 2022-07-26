//go:build windows

package scanner

import (
	"path/filepath"
	"syscall"
)

func isHiddenFile(filename string) (bool, error) {
	base := filepath.Base(filename)
	if base[0:1] == "." {
		return true, nil
	}

	pointer, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
