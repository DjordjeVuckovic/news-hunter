package pg

import (
	"context"
	"encoding/json"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader struct {
	db *pgxpool.Pool
}

func NewReader(pool *ConnectionPool) (*Reader, error) {
	return &Reader{db: pool.conn}, nil
}

func (r *Reader) SearchBasic(ctx context.Context, query string, page int, size int) ([]domain.Article, error) {
	sql := `
		SELECT id, title, subtitle, content, author, description, url, language, created_at, metadata
     	FROM articles
		WHERE to_tsvector('english', title || ' ' || subtitle || ' ' || content) @@ plainto_tsquery('english', $1)
		ORDER BY created_at DESC
		LIMIT $2 
		OFFSET $3;
     `
	dbRows, err := r.db.Query(ctx, sql, query, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()
	var articles []domain.Article
	for dbRows.Next() {
		var article domain.Article
		var metadataJSON []byte
		if err := dbRows.Scan(
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
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadataJSON, &article.Metadata); err != nil {
			return nil, err
		}

		articles = append(articles, article)
	}
	if err := dbRows.Err(); err != nil {
		return nil, err
	}
	return articles, nil
}
