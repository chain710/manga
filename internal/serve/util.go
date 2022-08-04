package serve

import (
	"database/sql"
	"errors"
	"github.com/chain710/manga/internal/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type jsonHandler func(context *gin.Context) (interface{}, error)
type imageHandler func(context *gin.Context) (*types.Image, error)

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
			context.Data(http.StatusOK, "image/"+strings.ToLower(result.Format), result.Data)
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
