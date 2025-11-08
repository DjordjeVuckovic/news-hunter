package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader struct {
	db *pgxpool.Pool
}

func NewReader(pool *ConnectionPool) (*Reader, error) {
	return &Reader{db: pool.conn}, nil
}

// SearchFullText implements storage.Reader interface
// Performs token-based full-text search using PostgreSQL's tsvector and plainto_tsquery
func (r *Reader) SearchFullText(ctx context.Context, query *domain.FullTextQuery, cursor *dto.Cursor, size int) (*storage.SearchResult, error) {
	slog.Info("Executing pg full-text search", "query", query.Text, "has_cursor", cursor != nil, "size", size)

	var globalMaxScore float64
	var count int64
	maxSQL := `
			SELECT COALESCE(MAX(ts_rank(search_vector, plainto_tsquery('english', $1))), 0.0) as max_score, COUNT(*)
			FROM articles
			WHERE search_vector @@ plainto_tsquery('english', $1)
		`
	if err := r.db.QueryRow(ctx, maxSQL, query.Text).Scan(&globalMaxScore, &count); err != nil || globalMaxScore <= 0 {
		slog.Error("Failed to fetch global max score", "error", err)
		return nil, fmt.Errorf("cannot fetch global max score: %w", err)
	}
	slog.Info("Computed global max score", "max_score", globalMaxScore, "total_matches", count)

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
		args = []interface{}{query.Text, size + 1}
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
		args = []interface{}{query.Text, cursor.Score, cursor.ID, size + 1}
	}

	rows, err := r.db.Query(ctx, searchSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var articles []dto.ArticleSearchResult
	var rawScores []float64

	for rows.Next() {
		var metadataJSON []byte
		var rawScore float64
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
			&rawScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &article.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		searchResult := dto.ArticleSearchResult{
			Article:         article,
			Score:           utils.RoundFloat64(rawScore, domain.ScoreDecimalPlaces),
			ScoreNormalized: utils.RoundFloat64(rawScore/globalMaxScore, domain.ScoreDecimalPlaces),
		}

		articles = append(articles, searchResult)
		rawScores = append(rawScores, rawScore)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	slog.Info("PG search results fetched",
		"total_page_matches", len(articles),
		"global_max_score", globalMaxScore)

	hasMore := len(articles) > size
	if hasMore {
		articles = articles[:size]
		rawScores = rawScores[:size]
	}

	var nextCursor *dto.Cursor
	if hasMore && len(articles) > 0 {
		nextCursor = &dto.Cursor{
			Score: rawScores[len(rawScores)-1],
			ID:    articles[len(articles)-1].Article.ID,
		}
	}

	return &storage.SearchResult{
		Hits:         articles,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
		MaxScore:     utils.RoundFloat64(globalMaxScore, domain.ScoreDecimalPlaces),
		PageMaxScore: utils.RoundFloat64(rawScores[0], domain.ScoreDecimalPlaces),
		TotalMatches: count,
	}, nil
}

// SearchBoolean implements storage.BooleanSearcher interface
// Performs boolean search using PostgreSQL's tsquery with AND (&), OR (|), NOT (!) operators
func (r *Reader) SearchBoolean(ctx context.Context, query *domain.BooleanQuery, cursor *dto.Cursor, size int) (*storage.SearchResult, error) {
	slog.Info("Executing pg boolean search", "expression", query.Expression, "has_cursor", cursor != nil, "size", size)

	// TODO: Implement boolean query parser
	// Parse query.Expression: "climate AND (change OR warming) AND NOT politics"
	// Convert to PostgreSQL tsquery syntax: "climate & (change | warming) & !politics"
	// Use websearch_to_tsquery or to_tsquery for parsing

	return nil, fmt.Errorf("boolean search not yet implemented for PostgreSQL")
}

// Compile-time interface assertions
var _ storage.Reader = (*Reader)(nil)
var _ storage.BooleanSearcher = (*Reader)(nil)
