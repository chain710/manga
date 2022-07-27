//go:build !windows

package parser

import "path/filepath"

func isHiddenFile(filename string) (bool, error) {
	base := filepath.Base(filename)
	return base[0:1] == ".", nil
}
