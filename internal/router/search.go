package router

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/labstack/echo/v4"
)

type SearchRouter struct {
	e       *echo.Echo
	storage storage.Storer
}
