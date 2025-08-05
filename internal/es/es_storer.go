package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/google/uuid"
)

type Storer struct {
	typedClient *elasticsearch.TypedClient
	indexName   string
	config      Config
}

type Config struct {
	Addresses []string
	IndexName string
	Username  string
	Password  string
}

// Document ESDocument represents the document structure for Elasticsearch
type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	Language    string    `json:"language"`
	CreatedAt   time.Time `json:"created_at"`
	SourceId    string    `json:"source_id"`
	SourceName  string    `json:"source_name"`
	PublishedAt time.Time `json:"published_at"`
	Category    string    `json:"category"`
	ImportedAt  time.Time `json:"imported_at"`
	IndexedAt   time.Time `json:"indexed_at"`
}

func NewStorer(ctx context.Context, config Config) (*Storer, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
	}

	if config.Username != "" && config.Password != "" {
		cfg.Username = config.Username
		cfg.Password = config.Password
	}

	client, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch typedClient: %w", err)
	}
	storer := &Storer{
		typedClient: client,
		indexName:   config.IndexName,
		config:      config,
	}

	if err := storer.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure index exists: %w", err)
	}

	return storer, nil
}

func (e *Storer) Save(ctx context.Context, article domain.Article) (uuid.UUID, error) {
	doc := e.articleToESDocument(article)

	res, err := e.typedClient.Index(e.indexName).Id(doc.ID).Document(doc).Do(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to index document: %w", err)
	}

	articleID, err := uuid.Parse(doc.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse article ID: %w", err)
	}

	slog.Info("document indexed successfully", "id", doc.ID, "index", e.indexName, "result", res.Result)
	return articleID, nil
}

func (e *Storer) SaveBulk(ctx context.Context, articles []domain.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// We need to create a regular ES typedClient from the typed typedClient for esutil
	// This is a limitation - esutil needs the regular typedClient
	cfg := elasticsearch.Config{
		Addresses: e.config.Addresses,
	}

	if e.config.Username != "" && e.config.Password != "" {
		cfg.Username = e.config.Username
		cfg.Password = e.config.Password
	}

	// Create bulk indexer
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         e.indexName,
		Client:        e.typedClient,
		NumWorkers:    4,
		FlushBytes:    5e+6, // 5MB
		FlushInterval: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create bulk indexer: %w", err)
	}

	// Track results
	var successful, failed int64

	// Add documents to bulk indexer
	for _, article := range articles {
		doc := e.articleToESDocument(article)

		docBytes, err := json.Marshal(doc)
		if err != nil {
			slog.Error("failed to marshal document", "error", err, "id", doc.ID)
			failed++
			continue
		}

		err = bi.Add(
			ctx,
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: doc.ID,
				Body:       bytes.NewReader(docBytes),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					successful++
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					failed++
					if err != nil {
						slog.Error("bulk index error", "error", err, "id", item.DocumentID)
					} else {
						slog.Error("bulk index error", "status", res.Status, "error", res.Error.Type, "reason", res.Error.Reason, "id", item.DocumentID)
					}
				},
			},
		)
		if err != nil {
			failed++
			slog.Error("failed to add document to bulk indexer", "error", err, "id", doc.ID)
		}
	}

	// Close the indexer and wait for completion
	if err := bi.Close(ctx); err != nil {
		return fmt.Errorf("failed to close bulk indexer: %w", err)
	}

	slog.Info("Bulk indexing completed",
		"successful", successful,
		"failed", failed,
		"total", len(articles),
		"index", e.indexName)

	if failed > 0 {
		return fmt.Errorf("failed to index %d out of %d articles", failed, len(articles))
	}

	return nil
}

func (e *Storer) articleToESDocument(article domain.Article) Document {
	if article.ID == uuid.Nil {
		article.ID = uuid.New()
	}
	return Document{
		ID:          article.ID.String(),
		Title:       article.Title,
		Subtitle:    article.Subtitle,
		Description: article.Description,
		Content:     article.Content,
		Author:      article.Author,
		URL:         article.URL.String(),
		Language:    article.Language,
		CreatedAt:   article.CreatedAt,
		SourceId:    article.Metadata.SourceId,
		SourceName:  article.Metadata.SourceName,
		PublishedAt: article.Metadata.PublishedAt,
		Category:    article.Metadata.Category,
		ImportedAt:  article.Metadata.ImportedAt,
		IndexedAt:   time.Now(),
	}
}

func (e *Storer) EnsureIndex(ctx context.Context) error {
	existsRes, err := e.typedClient.Indices.Exists(e.indexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if existsRes {
		slog.Info("Index already exists", "index", e.indexName)
		return nil
	}

	settings := types.IndexSettings{
		NumberOfShards:   "1",
		NumberOfReplicas: "0",
		Analysis: &types.IndexSettingsAnalysis{
			Analyzer: map[string]types.Analyzer{
				"multilingual_analyzer": types.StandardAnalyzer{
					Stopwords: []string{"_none_"},
				},
			},
		},
	}

	mappings := types.TypeMapping{
		Properties: map[string]types.Property{
			"id":           types.NewKeywordProperty(),
			"title":        e.createTextPropertyWithKeyword("multilingual_analyzer"),
			"subtitle":     e.createTextProperty("multilingual_analyzer"),
			"description":  e.createTextProperty("multilingual_analyzer"),
			"content":      e.createTextProperty("multilingual_analyzer"),
			"author":       e.createTextPropertyWithKeyword(""),
			"url":          types.NewKeywordProperty(),
			"language":     types.NewKeywordProperty(),
			"created_at":   types.NewDateProperty(),
			"source_id":    types.NewKeywordProperty(),
			"source_name":  e.createTextPropertyWithKeyword(""),
			"published_at": types.NewDateProperty(),
			"category":     types.NewKeywordProperty(),
			"imported_at":  types.NewDateProperty(),
			"indexed_at":   types.NewDateProperty(),
		},
	}

	createRes, err := e.typedClient.Indices.Create(e.indexName).
		Settings(&settings).
		Mappings(&mappings).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if !createRes.Acknowledged {
		return fmt.Errorf("index creation was not acknowledged")
	}

	slog.Info("Index created successfully", "index", e.indexName)
	return nil
}

func (e *Storer) createTextProperty(analyzer string) types.Property {
	textProp := types.NewTextProperty()
	if analyzer != "" {
		textProp.Analyzer = &analyzer
	}
	return textProp
}

func (e *Storer) createTextPropertyWithKeyword(analyzer string) types.Property {
	textProp := types.NewTextProperty()
	if analyzer != "" {
		textProp.Analyzer = &analyzer
	}
	textProp.Fields = map[string]types.Property{
		"keyword": types.NewKeywordProperty(),
	}
	return textProp
}
