package pg

import (
	"context"
	"fmt"
	"log/slog"

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

// SaveBulk upserts a batch of embeddings. It COPYs into a temporary staging
// table, then inserts only rows whose article_id exists (orphans are skipped and
// logged) with ON CONFLICT upsert, making the operation re-runnable.
func (e *Embedder) SaveBulk(ctx context.Context, article []*embedding.Vec) error {
	if len(article) == 0 {
		return nil
	}

	tx, err := e.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE _embed_stage (
			article_id uuid,
			model_name text,
			embedding  vector
		) ON COMMIT DROP
	`)
	if err != nil {
		return fmt.Errorf("failed to create staging table: %w", err)
	}

	rows := make([][]any, len(article))
	for i, ar := range article {
		rows[i] = []any{ar.ID, ar.Model, pgvector.NewVector(ar.Embedding)}
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"_embed_stage"},
		[]string{"article_id", "model_name", "embedding"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to copy embeddings to staging: %w", err)
	}

	tag, err := tx.Exec(ctx, `
		INSERT INTO article_embeddings (article_id, model_name, embedding)
		SELECT s.article_id, s.model_name, s.embedding
		FROM _embed_stage s
		JOIN articles a ON a.id = s.article_id
		ON CONFLICT (article_id, model_name) DO UPDATE
		SET embedding = EXCLUDED.embedding
	`)
	if err != nil {
		return fmt.Errorf("failed to upsert article embeddings: %w", err)
	}

	if skipped := int64(len(article)) - tag.RowsAffected(); skipped > 0 {
		slog.Warn("skipped embeddings with no matching article",
			"skipped", skipped,
			"upserted", tag.RowsAffected(),
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit embeddings: %w", err)
	}

	return nil
}
