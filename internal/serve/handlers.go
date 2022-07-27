package serve

import (
	"github.com/chain710/manga/internal/db"
	"github.com/gin-gonic/gin"
)

type jsonHandler func(context *gin.Context) (interface{}, error)

type handlers struct {
	config   Config
	database db.Interface
}

func (h handlers) registerRoutes(router *gin.Engine) {
	v1 := router.Group(h.config.GetBaseURI("/apis/v1"))
	v1.GET("/library", wrapJSONHandler(h.listLibraries))
}

func (h *handlers) listLibraries(ctx *gin.Context) (interface{}, error) {
	return h.database.ListLibraries(ctx)
}
