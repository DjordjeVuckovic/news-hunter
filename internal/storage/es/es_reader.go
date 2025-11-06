package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
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

func (r *Reader) SearchFullText(ctx context.Context, query string, cursor *dto.Cursor, size int) (*storage.SearchResult, error) {
	slog.Info("Executing es full-text search", "query", query, "has_cursor", cursor != nil, "size", size)

	searchReq := r.client.Search().
		Index(r.indexName).
		Query(&types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: []string{"title^3", "description^2", "content"},
			},
		}).
		Size(size + 1). // Fetch size+1 to determine hasMore
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
		slog.Error("Elasticsearch query failed", "error", err, "query", query, "cursor", cursor != nil)
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	// Get max score for normalization (0-1 range)
	maxScore := domain.NormalizeScore((*float64)(res.Hits.MaxScore))

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
		lastItem := articles[len(articles)-1]
		lastRawScore := rawScores[len(rawScores)-1]
		nextCursor = &dto.Cursor{
			Score: lastRawScore,
			ID:    lastItem.Article.ID,
		}
	}

	return &storage.SearchResult{
		Items:        articles,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
		MaxScore:     float64(*res.Hits.MaxScore),
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
		rawScores = append(rawScores, rawScore) // Keep raw score for cursor
	}

	return articles, rawScores, nil
}
