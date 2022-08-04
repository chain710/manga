package serve

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"hash/fnv"
	"net/http"
	"os"
	"time"
)

type jsonHandler func(context *gin.Context) (interface{}, error)
type imageHandler func(context *gin.Context) (*imageData, error)

func wrapJSONHandler(h jsonHandler) gin.HandlerFunc {
	return func(context *gin.Context) {
		result, err := h(context)
		response := JSONResponse{
			Data: result,
		}
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.Error = errNotFound.Error()
			} else {
				response.Error = err.Error()
			}
		}
		context.JSON(http.StatusOK, &response)
	}
}

func wrapImageHandler(h imageHandler) gin.HandlerFunc {
	return func(context *gin.Context) {
		result, err := h(context)
		if err != nil {
			context.Status(errorToStatus(err))
		} else {
			context.Data(http.StatusOK, "image/"+result.format, result.data)
		}
	}
}

func errorToStatus(err error) int {
	if errors.Is(err, errInvalidRequest) {
		return http.StatusBadRequest
	} else if errors.Is(err, errNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func fileHash(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	in := struct {
		Path    string    `json:"path"`
		Size    int64     `json:"size"`
		ModTime time.Time `json:"modtime"`
	}{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	data, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	h64 := fnv.New64()
	_, _ = h64.Write(data)
	return base62.EncodeToString(h64.Sum(nil)), nil
}
