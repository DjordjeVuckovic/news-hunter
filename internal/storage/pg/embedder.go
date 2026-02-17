package pg

import (
	"context"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type Embedder struct {
	db *pgxpool.Pool
}

func NewEmbedder(pool *ConnectionPool) *Embedder {
	return &Embedder{db: pool.GetConn()}
}

func (e *Embedder) Save(ctx context.Context, article *embedding.Vec) (uuid.UUID, error) {
	vec := pgvector.NewVector(article.Embedding)
	cmd := `
		INSERT INTO article_embeddings (article_id, model_name, embedding)
		VALUES ($1, $2, $3)
		ON CONFLICT (article_id, model_name) DO UPDATE
		SET embedding = EXCLUDED.embedding
		RETURNING id
	`
	var id uuid.UUID
	err := e.db.QueryRow(
		ctx,
		cmd,
		article.ID,
		article.Model,
		vec,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert article embedding: %w", err)
	}

	return id, nil
}

func (e *Embedder) SaveBulk(ctx context.Context, article []*embedding.Vec) error {
	rows := make([][]any, len(article))
	for i, ar := range article {
		vec := pgvector.NewVector(ar.Embedding)
		rows[i] = []any{ar.ID, ar.Model, vec}
	}

	_, err := e.db.CopyFrom(
		ctx,
		pgx.Identifier{"article_embeddings"},
		[]string{"article_id", "model_name", "embedding"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("failed to insert article embeddings: %w", err)
	}

	return nil
}
