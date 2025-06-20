package storage

import (
	"context"
	"fmt"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgStorer struct {
	db *pgxpool.Pool
}

func NewPgStorage(ctx context.Context, connStr string) (*PgStorer, error) {
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := dbpool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return &PgStorer{db: dbpool}, nil
}

func (s *PgStorer) Save(ctx context.Context, article domain.Article) (uuid.UUID, error) {
	query := `
        INSERT INTO articles (title, url, published_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING;
    `
	_, err := s.db.Exec(ctx, query,
		article.Title,
		article.URL,
		article.Metadata.PublishedAt,
	)

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to insert article: %w", err)
	}

	return article.ID, nil
}
