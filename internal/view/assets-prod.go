//go:build !dev
// +build !dev

package view

import (
	"embed"
	"io/fs"
)

//go:embed public
var publicRoot embed.FS
var Assets, _ = fs.Sub(publicRoot, "public")
