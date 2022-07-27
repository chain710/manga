package serve

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func wrapJSONHandler(h jsonHandler) gin.HandlerFunc {
	return func(context *gin.Context) {
		result, err := h(context)
		response := JSONResponse{
			Data: result,
		}
		if err != nil {
			response.Error = err.Error()
		}
		context.JSON(http.StatusOK, &response)
	}
}
