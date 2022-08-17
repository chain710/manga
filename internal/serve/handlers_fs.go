package serve

import (
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/types"
	"github.com/gin-gonic/gin"
	"os"
	"path"
)

func (h *handlers) apiListDirectory(ctx *gin.Context) (interface{}, error) {
	var queryParam struct {
		Path string `form:"path"`
	}
	if err := ctx.ShouldBindQuery(&queryParam); err != nil {
		log.Debugf("bind query error: %s", err)
		return nil, errInvalidRequest
	}

	entries, err := os.ReadDir(queryParam.Path)
	if err != nil {
		log.Errorf("list dir %s error: %s", queryParam.Path, err)
		return nil, err
	}

	var result types.Directory
	result.Parent = path.Dir(queryParam.Path)
	if result.Parent == queryParam.Path {
		result.Parent = ""
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		de := types.DirEntry{
			Name: entry.Name(),
			Type: types.EntryTypeDirectory,
			Path: path.Join(queryParam.Path, entry.Name()),
		}

		result.Entries = append(result.Entries, de)
	}

	return result, nil
}
