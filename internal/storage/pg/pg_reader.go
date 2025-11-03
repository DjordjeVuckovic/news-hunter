package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader struct {
	db *pgxpool.Pool
}

func NewReader(pool *ConnectionPool) (*Reader, error) {
	return &Reader{db: pool.conn}, nil
}

func (r *Reader) SearchFullText(ctx context.Context, query string, cursor *dto.Cursor, size int) (*storage.SearchResult, error) {
	slog.Info("Executing pg full-text search", "query", query, "has_cursor", cursor != nil, "size", size)

	var searchSQL string
	var args []interface{}

	if cursor == nil {
		searchSQL = `
			SELECT
				id, title, subtitle, content, author, description, url, language, created_at, metadata,
				ts_rank(search_vector, plainto_tsquery('english', $1)) as rank
			FROM articles
			WHERE search_vector @@ plainto_tsquery('english', $1)
			ORDER BY rank DESC, id DESC
			LIMIT $2
		`
		args = []interface{}{query, size + 1}
	} else {
		searchSQL = `
			SELECT
				id, title, subtitle, content, author, description, url, language, created_at, metadata,
				ts_rank(search_vector, plainto_tsquery('english', $1)) as rank
			FROM articles
			WHERE search_vector @@ plainto_tsquery('english', $1)
			  AND (ts_rank(search_vector, plainto_tsquery('english', $1)), id) < ($2, $3)
			ORDER BY rank DESC, id DESC
			LIMIT $4
		`
		args = []interface{}{query, cursor.Rank, cursor.ID, size + 1}
	}

	var err error

	rows, err := r.db.Query(ctx, searchSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var articles []dto.ArticleSearchResult
	for rows.Next() {
		var metadataJSON []byte
		var rank float32
		var article dto.Article

		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Subtitle,
			&article.Content,
			&article.Author,
			&article.Description,
			&article.URL,
			&article.Language,
			&article.CreatedAt,
			&metadataJSON,
			&rank,
		); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &article.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		searchResult := dto.ArticleSearchResult{
			Article: article,
			Rank:    rank,
		}

		articles = append(articles, searchResult)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Build result with domain cursor (no encoding)
	hasMore := len(articles) > size
	if hasMore {
		articles = articles[:size] // Trim to requested size
	}

	var nextCursor *dto.Cursor
	if hasMore && len(articles) > 0 {
		lastItem := articles[len(articles)-1]
		nextCursor = &dto.Cursor{
			Rank: lastItem.Rank,
			ID:   lastItem.Article.ID,
		}
	}

	return &storage.SearchResult{
		Items:      articles,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
