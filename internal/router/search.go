package router

import (
	"fmt"
	"log/slog"
	"strconv"

	dquery "github.com/DjordjeVuckovic/news-hunter/internal/domain/query"
	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/pkg/pagination"
	"github.com/labstack/echo/v4"
)

type SearchRouter struct {
	e       *echo.Echo
	storage storage.FTSSearcher
}

func NewSearchRouter(e *echo.Echo, storage storage.FTSSearcher) *SearchRouter {
	return &SearchRouter{
		e:       e,
		storage: storage,
	}
}

func (r *SearchRouter) Bind() {
	r.e.GET("/v1/articles/search", r.searchHandler)
}

// FTSSearchResponse represents the API response for full-text search
// This is a concrete type for Swagger documentation (swag doesn't support generics yet)
type FTSSearchResponse struct {
	NextCursor   *string                   `json:"next_cursor,omitempty"`
	HasMore      bool                      `json:"has_more"`
	MaxScore     float64                   `json:"max_score,omitempty"`
	PageMaxScore float64                   `json:"page_max_score,omitempty"`
	TotalMatches int64                     `json:"total_matches,omitempty"`
	Hits         []dto.ArticleSearchResult `json:"hits"`
}

// searchHandler handles full-text search requests with cursor-based pagination
// @Summary Full-text news search
// @Description Search for news articles using full-text query with cursor-based pagination
// @Tags search
// @Accept  json
// @Produce  json
// @Param query query string true "Search query"
// @Param cursor query string false "Cursor for pagination (base64-encoded)"
// @Param size query int false "Page size (default: 100, max: 10000)"
// @Success 200 {object} FTSSearchResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/articles/search [get]
func (r *SearchRouter) searchHandler(c echo.Context) error {
	query := c.QueryParam("query")
	cursorStr := c.QueryParam("cursor")
	sizeStr := c.QueryParam("size")

	if query == "" {
		return c.JSON(400, map[string]string{"error": "query parameter is required"})
	}

	sizeInt := pagination.PageDefaultSize
	if sizeStr != "" {
		var err error
		sizeInt, err = strconv.Atoi(sizeStr)
		if err != nil || sizeInt < 1 {
			return c.JSON(400, map[string]string{"error": "invalid size parameter"})
		}
		if sizeInt > pagination.PageMaxSize {
			return c.JSON(400,
				map[string]string{
					"error": fmt.Sprintf("size parameter exceeds maximum of %d", pagination.PageMaxSize),
				})
		}
	}

	var cursor *dto.Cursor
	if cursorStr != "" {
		var err error
		cursor, err = dto.DecodeCursor(cursorStr)
		if err != nil {
			return c.JSON(400, map[string]string{"error": "invalid cursor parameter"})
		}
	}

	fullTextQuery := dquery.NewFullTextQuery(query)
	searchResult, err := r.storage.SearchFullText(c.Request().Context(), fullTextQuery, cursor, sizeInt)
	if err != nil {
		slog.Error("Failed to execute full-text search", "error", err, "query", query)
		return c.JSON(500, map[string]string{"error": "internal server error"})
	}

	var nextCursorStr *string
	if searchResult.NextCursor != nil {
		encoded, err := dto.EncodeCursor(searchResult.NextCursor.Score, searchResult.NextCursor.ID)
		if err != nil {
			slog.Error("Failed to encode cursor", "error", err)
			return c.JSON(500, map[string]string{"error": "internal server error"})
		}
		nextCursorStr = &encoded
	}

	// Create API response with encoded cursor string
	apiResponse := FTSSearchResponse{
		Hits:         searchResult.Hits,
		NextCursor:   nextCursorStr,
		HasMore:      searchResult.HasMore,
		MaxScore:     searchResult.MaxScore,
		PageMaxScore: searchResult.PageMaxScore,
		TotalMatches: searchResult.TotalMatches,
	}

	return c.JSON(200, apiResponse)
}
