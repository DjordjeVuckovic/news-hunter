package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PgStorer struct {
	db *pgxpool.Pool
}

type PgStorerConfig struct {
	ConnStr string
}

func NewPgStorer(ctx context.Context, cfg PgStorerConfig) (*PgStorer, error) {
	dbpool, err := pgxpool.New(ctx, cfg.ConnStr)
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
		article.Language = domain.ArticleDefaultLanguage
	}
	if article.CreatedAt.IsZero() {
		article.CreatedAt = time.Now()
	}

	// Set ImportedAt if not already set
	if article.Metadata.ImportedAt.IsZero() {
		article.Metadata.ImportedAt = time.Now()
	}

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(article.Metadata)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	cmd := `
        INSERT INTO articles (id, title, subtitle, content, author, description, url, language, created_at, metadata)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id;
    `
	var id uuid.UUID
	err = s.db.QueryRow(
		ctx,
		cmd,
		article.ID,
		article.Title,
		article.Subtitle,
		article.Content,
		article.Author,
		article.Description,
		article.Language,
		article.CreatedAt,
		metadataJSON,
	).Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to insert article: %w", err)
	}

	return id, nil
}

func (s *PgStorer) SaveBulk(ctx context.Context, articles []domain.Article) error {
	rows := make([][]interface{}, len(articles))
	now := time.Now()

	for i, a := range articles {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		if a.Language == "" {
			a.Language = domain.ArticleDefaultLanguage
		}
		if a.CreatedAt.IsZero() {
			a.CreatedAt = now
		}

		// Set ImportedAt if not already set
		if a.Metadata.ImportedAt.IsZero() {
			a.Metadata.ImportedAt = now
		}

		// Marshal metadata to JSON
		metadataJSON, err := json.Marshal(a.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata for article %d: %w", i, err)
		}

		rows[i] = []interface{}{
			a.ID,
			a.Title,
			a.Subtitle,
			a.Content,
			a.Author,
			a.Description,
			a.URL.String(),
			a.Language,
			a.CreatedAt,
			metadataJSON,
		}
	}

	_, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{"articles"},
		[]string{"id", "title", "subtitle", "content", "author", "description", "url", "language", "created_at", "metadata"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("failed to bulk insert articles: %w", err)
	}
	return nil
}
