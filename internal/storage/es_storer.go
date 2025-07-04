package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type EsStorer struct {
	client    *elasticsearch.Client
	indexName string
}

type EsStorerConfig struct {
	Addresses []string
	IndexName string
	Username  string
	Password  string
}

// ESDocument represents the document structure for Elasticsearch
type ESDocument struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Content     string                 `json:"content"`
	Author      string                 `json:"author"`
	URL         string                 `json:"url"`
	URLToImage  string                 `json:"url_to_image"`
	PublishedAt time.Time              `json:"published_at"`
	Source      string                 `json:"source"`
	Category    string                 `json:"category"`
	Language    string                 `json:"language"`
	Country     string                 `json:"country"`
	Metadata    map[string]interface{} `json:"metadata"`
	IndexedAt   time.Time              `json:"indexed_at"`
}

func NewEsStorer(ctx context.Context, config EsStorerConfig) (*EsStorer, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
	}

	if config.Username != "" && config.Password != "" {
		cfg.Username = config.Username
		cfg.Password = config.Password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	storer := &EsStorer{
		client:    client,
		indexName: config.IndexName,
	}
	// Create index if it doesn't exist
	if err := storer.ensureIndex(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure index exists: %w", err)
	}

	return storer, nil
}

func (e *EsStorer) Save(ctx context.Context, article domain.Article) (uuid.UUID, error) {
	doc := e.articleToESDocument(article)

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      e.indexName,
		DocumentID: doc.ID,
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to index document: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(res.Body)

	if res.IsError() {
		return uuid.Nil, fmt.Errorf("error indexing document: %s", res.String())
	}

	articleID, err := uuid.Parse(doc.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse article ID: %w", err)
	}

	slog.Info("document indexed successfully", "id", doc.ID, "index", e.indexName)
	return articleID, nil
}

func (e *EsStorer) SaveBulk(ctx context.Context, articles []domain.Article) error {
	if len(articles) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, article := range articles {
		doc := e.articleToESDocument(article)

		// Bulk index action
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": e.indexName,
				"_id":    doc.ID,
			},
		}

		actionJSON, err := json.Marshal(action)
		if err != nil {
			return fmt.Errorf("failed to marshal bulk action: %w", err)
		}

		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}

		buf.Write(actionJSON)
		buf.WriteByte('\n')
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("error in bulk request: %s", res.String())
	}

	// Parse bulk response to check for individual errors
	var bulkRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		return fmt.Errorf("failed to parse bulk response: %w", err)
	}

	if errors, ok := bulkRes["errors"].(bool); ok && errors {
		slog.Warn("Some documents failed to index in bulk operation")
	}

	slog.Info("Bulk indexing completed", "count", len(articles), "index", e.indexName)
	return nil
}

func (e *EsStorer) articleToESDocument(article domain.Article) ESDocument {
	// Convert metadata map

	return ESDocument{
		ID:          article.ID.String(),
		Title:       article.Title,
		Description: article.Description,
		Content:     article.Content,
		Author:      article.Author,
		URL:         article.URL.String(),
		Language:    article.Language,
		IndexedAt:   time.Now(),
	}
}

func (e *EsStorer) ensureIndex(ctx context.Context) error {
	req := esapi.IndicesExistsRequest{
		Index: []string{e.indexName},
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(res.Body)

	if res.StatusCode == 200 {
		slog.Info("Index already exists", "index", e.indexName)
		return nil
	}

	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"multilingual_analyzer": {
						"type": "standard",
						"stopwords": "_none_"
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"title": {
					"type": "text",
					"analyzer": "multilingual_analyzer",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"description": {
					"type": "text",
					"analyzer": "multilingual_analyzer"
				},
				"content": {
					"type": "text",
					"analyzer": "multilingual_analyzer"
				},
				"author": {
					"type": "text",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"url": {"type": "keyword"},
				"url_to_image": {"type": "keyword"},
				"published_at": {"type": "date"},
				"source": {
					"type": "text",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"category": {"type": "keyword"},
				"language": {"type": "keyword"},
				"country": {"type": "keyword"},
				"metadata": {"type": "object"},
				"indexed_at": {"type": "date"}
			}
		}
	}`

	createReq := esapi.IndicesCreateRequest{
		Index: e.indexName,
		Body:  strings.NewReader(mapping),
	}

	createRes, err := createReq.Do(ctx, e.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(createRes.Body)

	if createRes.IsError() {
		return fmt.Errorf("error creating index: %s", createRes.String())
	}

	slog.Info("Index created successfully", "index", e.indexName)
	return nil
}
