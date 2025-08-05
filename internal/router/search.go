package router

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/labstack/echo/v4"
	"strconv"
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
	r.e.GET("/search", r.searchHandler)
}

func (r *SearchRouter) searchHandler(c echo.Context) error {
	query := c.QueryParam("query")
	page := c.QueryParam("page")
	size := c.QueryParam("size")

	// Validate query, page, and size parameters
	if query == "" {
		return c.JSON(400, map[string]string{"error": "query parameter is required"})
	}

	// Convert page and size to integers with default values
	pageInt := 1
	sizeInt := 10

	if page != "" {
		var err error
		pageInt, err = strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			pageInt = 1
		}
	}

	if size != "" {
		var err error
		sizeInt, err = strconv.Atoi(size)
		if err != nil || sizeInt < 1 {
			sizeInt = 10
		}
	}

	results, err := r.storage.SearchBasic(c.Request().Context(), query, pageInt, sizeInt)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "internal server error"})
	}

	return c.JSON(200, results)
}
