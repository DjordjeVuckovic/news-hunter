package storage

import (
	"context"
	"fmt"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgStorer struct {
	db *pgxpool.Pool
}

func NewPgStorer(ctx context.Context, connStr string) (*PgStorer, error) {
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
	if article.ID == uuid.Nil {
		article.ID = uuid.New()
	}
	if article.Language == "" {
		article.Language = "english"
	}

	cmd := `
        INSERT INTO articles (id, title, subtitle, content, author, language)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO NOTHING
        RETURNING id;
    `
	var id uuid.UUID
	err := s.db.QueryRow(
		ctx,
		cmd,
		article.ID,
		article.Title,
		article.Subtitle,
		article.Content,
		article.Author,
		article.Language,
	).Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to insert article: %w", err)
	}

	return id, nil
}

func (s *PgStorer) SaveBulk(ctx context.Context, articles []domain.Article) error {
	rows := make([][]interface{}, len(articles))
	for i, a := range articles {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		if a.Language == "" {
			a.Language = "english"
		}
		rows[i] = []interface{}{
			a.ID,
			a.Title,
			a.Subtitle,
			a.Content,
			a.Author,
			a.Language,
		}
	}

	_, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{"articles"},
		[]string{"id", "title", "subtitle", "content", "author", "language"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("failed to bulk insert articles: %w", err)
	}
	return nil
}
