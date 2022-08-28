package static

import (
	"embed"
	"github.com/chain710/manga/internal/log"
	ginstatic "github.com/gin-contrib/static"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist
var publicRoot embed.FS
var FS = newStaticFS(publicRoot, "dist")

type staticFS struct {
	http.FileSystem
}

var _ ginstatic.ServeFileSystem = staticFS{}

func (s staticFS) Exists(prefix string, path string) bool {
	trimPath := "/" + strings.TrimLeft(strings.TrimPrefix(path, prefix), "/")
	f, err := s.Open(trimPath)
	if err != nil {
		return false
	}

	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		log.Errorf("unexpect file (%s) stat error: %s", trimPath, err)
	}
	if strings.HasSuffix(path, "/") {
		return info.IsDir()
	} else {
		return !info.IsDir()
	}
}

func newStaticFS(efs embed.FS, sub string) ginstatic.ServeFileSystem {
	subFS, err := fs.Sub(efs, sub)
	if err != nil {
		log.Panicf("sub fs %s error: %s", sub, err)
	}

	return staticFS{FileSystem: http.FS(subFS)}
}
