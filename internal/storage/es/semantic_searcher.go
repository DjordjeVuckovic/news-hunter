package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/api/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	dquery "github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/google/uuid"
)

// minNumCandidates is the floor for the kNN num_candidates parameter; a larger
// candidate pool improves recall at the cost of latency.
const minNumCandidates = 100

// SemanticSearcher is the Elasticsearch implementation of
// storage.SemanticSearcher. It embeds the query with the same model the stored
// document vectors were produced with and runs an approximate kNN search over
// the dense_vector "embedding" field written by Embedder.
type SemanticSearcher struct {
	client    *elasticsearch.TypedClient
	indexName string
	embedder  *embedding.Embedder
	model     string
}

func NewSemanticSearcher(config ClientConfig, embedder *embedding.Embedder, model string) (*SemanticSearcher, error) {
	client, err := newClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}
	return &SemanticSearcher{
		client:    client,
		indexName: config.IndexName,
		embedder:  embedder,
		model:     model,
	}, nil
}

func (s *SemanticSearcher) SearchSemantic(ctx context.Context, query *dquery.Semantic, baseOpts *dquery.BaseOptions) (*storage.VectorSearchResult, error) {
	vec, err := s.embedder.EmbedQuery(ctx, query.Query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	size := baseOpts.Size
	k := size + 1
	numCandidates := k * 10
	if numCandidates < minNumCandidates {
		numCandidates = minNumCandidates
	}

	knn := types.KnnSearch{
		Field:         "embedding",
		QueryVector:   vec.Embedding,
		K:             &k,
		NumCandidates: &numCandidates,
	}

	// PG filters by cosine distance (< threshold); ES kNN filters by minimum
	// cosine similarity. similarity = 1 - distance, so a distance threshold maps
	// to a similarity floor. Only apply when it yields a valid [0,1] similarity.
	if query.Threshold > 0 && query.Threshold <= 1 {
		sim := float32(1 - query.Threshold)
		knn.Similarity = &sim
	}

	slog.Info("Executing es semantic kNN search",
		"query", query.Query,
		"model", s.model,
		"k", k,
		"num_candidates", numCandidates,
		"size", size)

	res, err := s.client.Search().
		Index(s.indexName).
		Knn(knn).
		SourceIncludes_(
			"id", "title", "subtitle", "content", "description",
			"author", "url", "language", "created_at",
			"source_id", "source_name", "published_at", "category", "imported_at",
		).
		Size(size + 1).
		Do(ctx)
	if err != nil {
		slog.Error("Elasticsearch kNN query failed", "error", err, "query", query.Query)
		return nil, fmt.Errorf("failed to execute semantic search: %w", err)
	}

	hits, err := s.mapToArticles(res.Hits.Hits)
	if err != nil {
		return nil, fmt.Errorf("failed to map semantic search results: %w", err)
	}

	slog.Info("ES semantic search results fetched",
		"total_matches", res.Hits.Total.Value,
		"returned_count", len(hits))

	hasMore := len(hits) > size
	if hasMore {
		hits = hits[:size]
	}

	var nextCursor *dquery.Cursor
	if hasMore && len(hits) > 0 {
		nextCursor = &dquery.Cursor{
			ID: hits[len(hits)-1].ID,
		}
	}

	return &storage.VectorSearchResult{
		Hits:       hits,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *SemanticSearcher) mapToArticles(hits []types.Hit) ([]dto.Article, error) {
	articles := make([]dto.Article, 0, len(hits))
	for _, hit := range hits {
		var doc ArticleDocument
		if err := json.Unmarshal(hit.Source_, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}

		articles = append(articles, dto.Article{
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
		})
	}
	return articles, nil
}

// Compile-time interface assertion
var _ storage.SemanticSearcher = (*SemanticSearcher)(nil)
