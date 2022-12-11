package serve

import (
	"database/sql"
	"errors"
	"github.com/chain710/manga/internal/cache"
	"github.com/chain710/manga/internal/log"
	"github.com/chain710/manga/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"image"
	"modernc.org/mathutil"
	"net/http"
	"strconv"
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
		status := http.StatusOK
		if err != nil {
			if errors.Is(err, errInvalidRequest) {
				status = http.StatusBadRequest
				response.Error = err.Error()
			} else if errors.Is(err, errNotFound) || errors.Is(err, sql.ErrNoRows) {
				status = http.StatusNotFound
				response.Error = errNotFound.Error()
			} else {
				status = http.StatusInternalServerError
				response.Error = err.Error()
			}
		}
		context.JSON(status, &response)
	}
}

func wrapImageHandler(h imageHandler, c *cache.Images) gin.HandlerFunc {
	return func(context *gin.Context) {
		key := context.Request.URL.Path
		if c != nil {
			// try cache first
			if img, ok := c.Get(key); ok {
				log.Debugf("hit image cache for %s", key)
				if match := context.GetHeader("If-None-Match"); match != "" {
					if strings.Contains(match, img.Hash) {
						context.Status(http.StatusNotModified)
						return
					}
				}
				writeImage(context, &img)
				return
			}
		}

		// call real function get image
		result, err := h(context)
		if err != nil {
			context.Status(errorToStatus(err))
		} else {
			writeImage(context, result)
			if c != nil {
				c.Set(key, *result)
				log.Debugf("update image cache for %s", key)
			}
		}
	}
}

func writeImage(context *gin.Context, img *types.Image) {
	context.Header("Cache-Control", "public, max-age=31536000")
	context.Header("ETag", img.Hash)
	context.Data(http.StatusOK, "image/"+strings.ToLower(img.Format), img.Data)
}

func errorToStatus(err error) int {
	if errors.Is(err, errInvalidRequest) {
		return http.StatusBadRequest
	} else if errors.Is(err, errNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func parseRect(r string) (image.Rectangle, error) {
	const maxDimension = 1000
	if len(r) != 12 {
		return image.Rectangle{}, errors.New("invalid rect string, len should be 12")
	}

	var merr *multierror.Error
	left, err := strconv.ParseInt(r[0:3], 16, 32)
	merr = multierror.Append(merr, err)
	top, err := strconv.ParseInt(r[3:6], 16, 32)
	merr = multierror.Append(merr, err)
	width, err := strconv.ParseInt(r[6:9], 16, 32)
	merr = multierror.Append(merr, err)
	height, err := strconv.ParseInt(r[9:12], 16, 32)
	merr = multierror.Append(merr, err)
	if top < 0 {
		merr = multierror.Append(merr, errors.New("top < 0"))
	}
	if left < 0 {
		merr = multierror.Append(merr, errors.New("left < 0"))
	}

	if width <= 0 {
		merr = multierror.Append(merr, errors.New("width <= 0"))
	}

	if height <= 0 {
		merr = multierror.Append(merr, errors.New("height <= 0"))
	}

	if width+left > maxDimension {
		merr = multierror.Append(merr, errors.New("exceed width limit"))
	}

	if height+top > maxDimension {
		merr = multierror.Append(merr, errors.New("exceed height limit"))
	}

	if err := merr.ErrorOrNil(); err != nil {
		return image.Rectangle{}, err
	}
	rect := image.Rect(int(left), int(top), int(width+left), int(height+top))
	return rect, nil
}

func bestFit(x, y, fitX, fitY int) (int, int) {
	if x == 0 || y == 0 || fitX == 0 || fitY == 0 {
		return 0, 0
	}

	aspect := float64(x) / float64(y)
	fitAspect := float64(fitX) / float64(fitY)
	if aspect == fitAspect {
		return mathutil.Min(x, fitX), mathutil.Min(y, fitY)
	} else if aspect > fitAspect {
		rx := mathutil.Min(x, fitX)
		ry := int(float64(rx) / aspect)
		return rx, ry
	} else {
		ry := mathutil.Min(y, fitY)
		rx := int(float64(ry) * aspect)
		return rx, ry
	}
}

func sanitizeTextSearchQuery(a string) string {
	if a == "" {
		return ""
	}

	r := strings.NewReplacer("&", " ")
	b := r.Replace(a)

	strArray := strings.Split(b, " ")
	newStrArray := make([]string, 0)
	for _, str := range strArray {
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			newStrArray = append(newStrArray, str)
		}
	}

	return strings.Join(newStrArray, " or ")
}
