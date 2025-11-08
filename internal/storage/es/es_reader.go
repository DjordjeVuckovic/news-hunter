package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/pkg/utils"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/google/uuid"
)

type Reader struct {
	client    *elasticsearch.TypedClient
	indexName string
}

func NewReader(config ClientConfig) (*Reader, error) {
	client, err := newClient(config)

	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	return &Reader{
		client:    client,
		indexName: config.IndexName,
	}, nil
}

// SearchLexical implements storage.Reader interface
// Performs token-based full-text search using Elasticsearch's multi_match query with BM25
func (r *Reader) SearchLexical(ctx context.Context, query *domain.LexicalQuery, cursor *dto.Cursor, size int) (*storage.SearchResult, error) {
	slog.Info("Executing es lexical search", "query", query.Text, "has_cursor", cursor != nil, "size", size)

	searchReq := r.client.Search().
		Index(r.indexName).
		Query(&types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query.Text,
				Fields: []string{"title", "description", "content"},
			},
		}).
		Size(size + 1).
		TrackScores(true)

	if cursor != nil {
		searchReq = searchReq.SearchAfter(
			types.FieldValue(cursor.Score),
			types.FieldValue(cursor.ID.String()),
		)
	}

	sortOrderDesc := sortorder.Desc
	searchReq = searchReq.Sort(
		&types.SortOptions{
			SortOptions: map[string]types.FieldSort{
				"_score": {Order: &sortOrderDesc},
			},
		},
		&types.SortOptions{
			SortOptions: map[string]types.FieldSort{
				"id": {Order: &sortOrderDesc},
			},
		},
	)

	var err error

	res, err := searchReq.Do(ctx)
	if err != nil {
		slog.Error("Elasticsearch query failed", "error", err, "query", query.Text, "cursor", cursor != nil)
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	maxScore := domain.CalcSafeScore((*float64)(res.Hits.MaxScore))

	articles, rawScores, err := r.mapToDomain(res.Hits.Hits, maxScore)
	if err != nil {
		return nil, fmt.Errorf("failed to map search results to domain: %w", err)
	}

	slog.Info("Es search results fetched",
		"total_matches", res.Hits.Total.Value,
		"returned_count", len(articles),
		"max_score", *res.Hits.MaxScore,
		"normalized_max", maxScore)

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
		MaxScore:     utils.RoundFloat64(float64(*res.Hits.MaxScore), domain.ScoreDecimalPlaces),
		PageMaxScore: utils.RoundFloat64(rawScores[0], domain.ScoreDecimalPlaces),
		TotalMatches: res.Hits.Total.Value,
	}, nil
}

func (r *Reader) mapToDomain(hits []types.Hit, maxScore float64) ([]dto.ArticleSearchResult, []float64, error) {
	if hits == nil {
		return make([]dto.ArticleSearchResult, 0), make([]float64, 0), nil
	}

	var articles []dto.ArticleSearchResult
	var rawScores []float64

	for _, hit := range hits {
		var doc Document
		if err := json.Unmarshal(hit.Source_, &doc); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}

		article := dto.Article{
			ID:          uuid.MustParse(doc.ID),
			Title:       doc.Title,
			Subtitle:    doc.Subtitle,
			Content:     doc.Content,
			Author:      doc.Author,
			Description: doc.Description,
			URL:         doc.URL,
			Language:    doc.Language,
			CreatedAt:   doc.CreatedAt,
			Metadata: dto.ArticleMetadata{
				SourceId:    doc.SourceId,
				SourceName:  doc.SourceName,
				PublishedAt: doc.PublishedAt,
				Category:    doc.Category,
				ImportedAt:  doc.ImportedAt,
			},
		}

		rawScore := float64(*hit.Score_)
		normalizedRank := rawScore / maxScore

		searchResult := dto.ArticleSearchResult{
			Article:         article,
			ScoreNormalized: normalizedRank,
			Score:           float64(*hit.Score_),
		}

		articles = append(articles, searchResult)
		rawScores = append(rawScores, rawScore)
	}

	return articles, rawScores, nil
}

// SearchBoolean implements storage.BooleanSearcher interface
// Performs boolean search using Elasticsearch's bool query with must, should, must_not clauses
func (r *Reader) SearchBoolean(ctx context.Context, query *domain.BooleanQuery, cursor *dto.Cursor, size int) (*storage.SearchResult, error) {
	slog.Info("Executing es boolean search", "expression", query.Expression, "has_cursor", cursor != nil, "size", size)

	// TODO: Implement boolean query parser
	// Parse query.Expression: "climate AND (change OR warming) AND NOT politics"
	// Convert to Elasticsearch bool query with must, should, must_not clauses

	return nil, fmt.Errorf("boolean search not yet implemented for Elasticsearch")
}

// Compile-time interface assertions
var _ storage.Reader = (*Reader)(nil)
var _ storage.BooleanSearcher = (*Reader)(nil)
