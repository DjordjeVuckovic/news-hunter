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

func (r *Reader) SearchBasic(ctx context.Context, query string, page int, size int) (*storage.SearchResult, error) {
	slog.Info("Executing es basic search", "query", query, "page", page, "size", size)

	from := (page - 1) * size

	res, err := r.client.Search().
		Index(r.indexName).
		Query(&types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  query,
				Fields: []string{"title^3", "description^2", "content"},
			},
		}).
		From(from).
		Size(size).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	var articles []dto.ArticleSearchResult
	if res.Hits.Hits != nil {
		for _, hit := range res.Hits.Hits {
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
	}

	total := int64(0)
	if res.Hits.Total != nil {
		total = res.Hits.Total.Value
	}

	hasMore := int64(from+size) < total

	return &storage.SearchResult{
		Items:   articles,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}, nil
}
