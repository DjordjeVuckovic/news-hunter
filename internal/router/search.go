package router

import (
	"strconv"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/labstack/echo/v4"
)

const (
	defaultPageSize = 10
	defaultPage     = 1
)

type SearchRouter struct {
	e       *echo.Echo
	storage storage.Reader
}

func NewSearchRouter(e *echo.Echo, storage storage.Reader) *SearchRouter {
	return &SearchRouter{
		e:       e,
		storage: storage,
	}
}

func (r *SearchRouter) Bind() {
	r.e.GET("/search/basic", r.searchHandler)
}

// searchHandler handles basic search requests
// @Summary Basic news search
// @Description Search for news articles using basic text query
// @Tags search
// @Accept  json
// @Produce  json
// @Param query query string true "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param size query int false "Page size (default: 10)"
// @Success 200 {object} storage.SearchResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /search/basic [get]
func (r *SearchRouter) searchHandler(c echo.Context) error {
	query := c.QueryParam("query")
	page := c.QueryParam("page")
	size := c.QueryParam("size")

	// Validate query, page, and size parameters
	if query == "" {
		return c.JSON(400, map[string]string{"error": "query parameter is required"})
	}

	pageInt := defaultPage
	sizeInt := defaultPageSize

	if page != "" {
		var err error
		pageInt, err = strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			pageInt = defaultPage
		}
	}

	if size != "" {
		var err error
		sizeInt, err = strconv.Atoi(size)
		if err != nil || sizeInt < 1 {
			sizeInt = defaultPageSize
		}
	}

	results, err := r.storage.SearchBasic(c.Request().Context(), query, pageInt, sizeInt)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "internal server error"})
	}

	return c.JSON(200, results)
}
