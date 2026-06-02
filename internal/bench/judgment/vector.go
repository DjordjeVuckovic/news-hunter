package judgment

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/document"
)

// defaultEmbeddingModel mirrors embedding.defaultModel (unexported there) so the
// vector judge can report an accurate model id when none is configured.
const defaultEmbeddingModel = "qwen3-embedding:0.6b"

// Embedder is the subset of *embedding.Embedder the vector/hybrid judges need.
// Declaring it here keeps the scoring logic testable with a fake backend.
type Embedder interface {
	EmbedQuery(ctx context.Context, query string) (*embedding.Vec, error)
	EmbedDocs(ctx context.Context, docs []document.Article) ([]embedding.Vec, error)
}

// VectorStrategy grades candidates by cosine similarity between the query
// embedding and each document embedding. Unlike lexical/bm25 it captures
// semantic relevance — paraphrases and related concepts with no shared
// keywords — which is what the semantic and hybrid tracks need. It requires an
// embedding backend (ollama) configured via StrategyOptions.
type VectorStrategy struct {
	embedder Embedder
	model    string
}

func NewVectorStrategy(opts StrategyOptions) (*VectorStrategy, error) {
	emb, model, err := embedderFromOptions(opts)
	if err != nil {
		return nil, err
	}
	return &VectorStrategy{embedder: emb, model: model}, nil
}

// NewVectorStrategyWithEmbedder injects an embedder directly (used in tests).
func NewVectorStrategyWithEmbedder(e Embedder, model string) *VectorStrategy {
	return &VectorStrategy{embedder: e, model: model}
}

func (VectorStrategy) Name() string { return string(StrategyVector) }

// ModelID lets cmd_judge stamp meta.JudgeModel with the embedding model used.
func (s VectorStrategy) ModelID() string {
	if s.model == "" {
		return string(StrategyVector)
	}
	return string(StrategyVector) + ":" + s.model
}

func (VectorStrategy) PreferredBatchSize() int { return poolBatchSize }

func (s VectorStrategy) GradeBatch(ctx context.Context, q GradingQuery, docs []GradingDoc) ([]GradedDoc, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	sims, err := s.similarities(ctx, q, docs)
	if err != nil {
		return nil, err
	}
	out := make([]GradedDoc, len(docs))
	for i, d := range docs {
		out[i] = GradedDoc{DocID: d.ID, Grade: gradeFromCosine(sims[i])}
	}
	return out, nil
}

func (s VectorStrategy) Grade(ctx context.Context, q GradingQuery, doc GradingDoc) (int, error) {
	sims, err := s.similarities(ctx, q, []GradingDoc{doc})
	if err != nil {
		return GradeUnjudged, err
	}
	return gradeFromCosine(sims[0]), nil
}

// similarities embeds the query once and all docs in one batch, returning the
// cosine similarity per doc in the same order as docs.
func (s VectorStrategy) similarities(ctx context.Context, q GradingQuery, docs []GradingDoc) ([]float64, error) {
	if s.embedder == nil {
		return nil, fmt.Errorf("vector strategy: no embedder configured")
	}
	qVec, err := s.embedder.EmbedQuery(ctx, q.Description)
	if err != nil {
		return nil, fmt.Errorf("embed query %q: %w", q.ID, err)
	}

	articles := make([]document.Article, len(docs))
	for i, d := range docs {
		articles[i] = document.Article{
			ID:      d.ID,
			Title:   d.Title,
			Content: strings.TrimSpace(d.Description + " " + d.Content),
		}
	}
	docVecs, err := s.embedder.EmbedDocs(ctx, articles)
	if err != nil {
		return nil, fmt.Errorf("embed docs: %w", err)
	}
	if len(docVecs) != len(docs) {
		return nil, fmt.Errorf("vector strategy: expected %d doc embeddings, got %d", len(docs), len(docVecs))
	}

	sims := make([]float64, len(docs))
	for i := range docs {
		sims[i] = cosine(qVec.Embedding, docVecs[i].Embedding)
	}
	return sims, nil
}

func embedderFromOptions(opts StrategyOptions) (Embedder, string, error) {
	if opts.EmbeddingBaseURL == "" {
		return nil, "", fmt.Errorf(
			"%s/%s strategy requires an embedding endpoint: set --embedding-base or EMBEDDING_BASE_URL (e.g. http://localhost:11434)",
			StrategyVector, StrategyHybrid)
	}
	client, err := embedding.NewOllamaClient(opts.EmbeddingBaseURL)
	if err != nil {
		return nil, "", fmt.Errorf("embedding client: %w", err)
	}
	model := opts.EmbeddingModel
	var eopts []embedding.EmbedderOption
	if model != "" {
		eopts = append(eopts, embedding.WithExecutorModel(model))
	} else {
		model = defaultEmbeddingModel
	}
	if opts.EmbeddingMaxLen != nil {
		eopts = append(eopts, embedding.WithExecutorMaxLength(*opts.EmbeddingMaxLen))
	}
	return embedding.NewEmbedder(client, eopts...), model, nil
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// gradeFromCosine maps cosine similarity to a grade. Thresholds suit
// sentence-embedding models (e.g. qwen3-embedding) and may need tuning per
// model — they are intentionally conservative.
func gradeFromCosine(c float64) int {
	switch {
	case c >= 0.70:
		return GradeHighly
	case c >= 0.55:
		return GradeRelevant
	case c >= 0.40:
		return GradeMarginally
	default:
		return GradeNotRelev
	}
}
