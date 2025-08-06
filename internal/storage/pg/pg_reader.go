package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader struct {
	db *pgxpool.Pool
}

func NewReader(pool *ConnectionPool) (*Reader, error) {
	return &Reader{db: pool.conn}, nil
}

func (r *Reader) SearchBasic(ctx context.Context, query string, page int, size int) (*storage.SearchResult, error) {

	offset := (page - 1) * size

	// Count total results
	countSQL := `
		SELECT COUNT(*)
		FROM articles
		WHERE search_vector @@ plainto_tsquery('english', $1)
	`

	var total int64
	err := r.db.QueryRow(ctx, countSQL, query).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	searchSQL := `
		SELECT 
			id, title, subtitle, content, author, description, url, language, created_at, metadata,
			ts_rank(search_vector, plainto_tsquery('english', $1)) as rank
		FROM articles
		WHERE search_vector @@ plainto_tsquery('english', $1)
		ORDER BY rank DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, searchSQL, query, size, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var article domain.Article
		var metadataJSON []byte
		var rank float32

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

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	hasMore := int64(offset+size) < total

	return &storage.SearchResult{
		Articles: articles,
		Total:    total,
		Page:     page,
		Size:     size,
		HasMore:  hasMore,
	}, nil
}
