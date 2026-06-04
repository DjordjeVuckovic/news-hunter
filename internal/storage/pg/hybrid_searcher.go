package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/api/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	dquery "github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"github.com/DjordjeVuckovic/news-hunter/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// hybridCandidateDepth bounds how many candidates each leg contributes to the fusion.
const hybridCandidateDepth = 200

type HybridSearcher struct {
	embedder *embedding.Embedder
	db       *pgxpool.Pool
}

func NewHybridSearcher(embedder *embedding.Embedder, pool *ConnectionPool) *HybridSearcher {
	return &HybridSearcher{
		embedder: embedder,
		db:       pool.GetConn(),
	}
}

// rrfScore is the RRF contribution for a document at lexRank/vecRank; rank 0 means absent.
// score = 1/(k + lexRank) + 1/(k + vecRank)
func rrfScore(k, lexRank, vecRank int) float64 {
	var score float64
	if lexRank > 0 {
		score += 1.0 / float64(k+lexRank)
	}
	if vecRank > 0 {
		score += 1.0 / float64(k+vecRank)
	}
	return score
}

// SearchHybrid fuses a lexical FTS ranking with a vector ranking via RRF in SQL.
func (s *HybridSearcher) SearchHybrid(ctx context.Context, query *dquery.Hybrid, baseOpts *dquery.BaseOptions) (*storage.SearchResult, error) {
	size := baseOpts.Size
	lang := query.GetLanguage()
	k := query.GetK()

	slog.Info("Executing PG hybrid RRF search",
		"query", query.Query,
		"language", lang,
		"k", k,
		"size", size)

	vec, err := s.embedder.EmbedQuery(ctx, query.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed hybrid query: %w", err)
	}
	vecEncoded := pgvector.NewVector(vec.Embedding)

	cmd := fmt.Sprintf(`
		WITH lexical AS (
			SELECT a.id AS article_id,
				   ROW_NUMBER() OVER (
					   ORDER BY ts_rank(a.search_vector, websearch_to_tsquery('%[1]s'::regconfig, $1)) DESC, a.id DESC
				   ) AS lex_rank
			FROM articles a
			WHERE a.search_vector @@ websearch_to_tsquery('%[1]s'::regconfig, $1)
			LIMIT $5
		),
		vector AS (
			SELECT e.article_id,
				   ROW_NUMBER() OVER (ORDER BY e.embedding <=> $2) AS vec_rank
			FROM article_embeddings e
			WHERE e.model_name = $3
			ORDER BY e.embedding <=> $2
			LIMIT $5
		),
		fused AS (
			SELECT COALESCE(l.article_id, v.article_id) AS article_id,
				   COALESCE(1.0 / ($4 + l.lex_rank), 0.0)
				 + COALESCE(1.0 / ($4 + v.vec_rank), 0.0) AS rrf_score
			FROM lexical l
			FULL OUTER JOIN vector v ON l.article_id = v.article_id
		)
		SELECT a.id, a.title, a.subtitle, a.content, a.author, a.description,
			   a.url, a.language, a.created_at, a.metadata, f.rrf_score
		FROM fused f
		INNER JOIN articles a ON a.id = f.article_id
		ORDER BY f.rrf_score DESC, a.id DESC
		LIMIT $6
	`, lang)

	rows, err := s.db.Query(
		ctx,
		cmd,
		query.Query,
		vecEncoded,
		vec.Model,
		k,
		hybridCandidateDepth,
		size+1,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute hybrid search query: %w", err)
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

		articles = append(articles, dto.ArticleSearchResult{
			Article: article,
			Score:   utils.RoundFloat64(rawScore, dquery.ScoreDecimalPlaces),
		})
		rawScores = append(rawScores, rawScore)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(articles) == 0 {
		return &storage.SearchResult{}, nil
	}

	hasMore := len(articles) > size
	if hasMore {
		articles = articles[:size]
		rawScores = rawScores[:size]
	}

	maxScore := rawScores[0]
	for i := range articles {
		articles[i].ScoreNormalized = utils.RoundFloat64(rawScores[i]/maxScore, dquery.ScoreDecimalPlaces)
	}

	var nextCursor *dquery.Cursor
	if hasMore {
		nextCursor = &dquery.Cursor{
			Score: rawScores[len(rawScores)-1],
			ID:    articles[len(articles)-1].Article.ID,
		}
	}

	slog.Info("PG hybrid search results fetched",
		"total_page_matches", len(articles),
		"max_score", maxScore)

	return &storage.SearchResult{
		Hits:         articles,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
		MaxScore:     utils.RoundFloat64(maxScore, dquery.ScoreDecimalPlaces),
		PageMaxScore: utils.RoundFloat64(maxScore, dquery.ScoreDecimalPlaces),
	}, nil
}

var _ storage.HybridSearcher = (*HybridSearcher)(nil)
