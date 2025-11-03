package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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

	// Build search request
	searchReq := r.client.Search().
		Index(r.indexName).
		Query(&types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: []string{"title^3", "description^2", "content"},
			},
		}).
		Size(size + 1) // Fetch size+1 to determine hasMore

	// Add search_after if cursor exists
	if cursor != nil {
		// search_after requires [score, id] in descending order
		searchReq = searchReq.SearchAfter(
			types.FieldValue(cursor.Rank),
			types.FieldValue(cursor.ID.String()),
		)
	}

	// Add sorting by score (desc) and id (desc) for stable pagination
	// Score is sorted DESC by default, we need to add id as tiebreaker
	// Note: We use the document's 'id' field (keyword), not '_id' (metadata field which is not sortable)
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

	articles, err := r.mapToDomain(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("failed to map search results to domain: %w", err)
	}

	slog.Info("Es search results fetched",
		"fetched_count", res.Hits.Total.Value,
		"max_score", res.Hits.MaxScore)

	hasMore := len(articles) > size
	if hasMore {
		articles = articles[:size] // Trim to requested size
	}

	var nextCursor *dto.Cursor
	if hasMore && len(articles) > 0 {
		lastItem := articles[len(articles)-1]
		nextCursor = &dto.Cursor{
			Rank: lastItem.Rank,
			ID:   lastItem.Article.ID,
		}
	}

	return &storage.SearchResult{
		Items:      articles,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (r *Reader) mapToDomain(hits []types.Hit) ([]dto.ArticleSearchResult, error) {
	if hits == nil {
		return make([]dto.ArticleSearchResult, 0), nil
	}

	var articles []dto.ArticleSearchResult
	for _, hit := range hits {
		var doc Document
		if err := json.Unmarshal(hit.Source_, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
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

		searchResult := dto.ArticleSearchResult{
			Article: article,
			Rank:    float32(*hit.Score_),
		}

		articles = append(articles, searchResult)
	}

	return articles, nil
}
