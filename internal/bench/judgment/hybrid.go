package judgment

import (
	"context"
	"fmt"
)

// Fusion weights for the hybrid score. Equal weighting is the common default
// for RRF-style hybrid retrieval; expose as constants so they're easy to tune.
const (
	hybridBM25Weight   = 0.5
	hybridVectorWeight = 0.5
)

// HybridStrategy fuses the BM25 lexical signal (pool-normalised) with vector
// cosine similarity (semantic signal), mirroring hybrid retrieval. The combined
// score is a weighted sum mapped to a grade 0-3. It judges the hybrid track on
// its own terms — neither pure keyword nor pure semantic relevance alone.
type HybridStrategy struct {
	bm25   BM25Strategy
	vector *VectorStrategy
}

func NewHybridStrategy(opts StrategyOptions) (*HybridStrategy, error) {
	v, err := NewVectorStrategy(opts)
	if err != nil {
		return nil, err
	}
	return &HybridStrategy{vector: v}, nil
}

// NewHybridStrategyWithEmbedder injects an embedder directly (used in tests).
func NewHybridStrategyWithEmbedder(e Embedder, model string) *HybridStrategy {
	return &HybridStrategy{vector: NewVectorStrategyWithEmbedder(e, model)}
}

func (HybridStrategy) Name() string { return string(StrategyHybrid) }

// ModelID stamps meta.JudgeModel with both components.
func (s HybridStrategy) ModelID() string {
	model := defaultEmbeddingModel
	if s.vector != nil && s.vector.model != "" {
		model = s.vector.model
	}
	return string(StrategyHybrid) + ":bm25+" + model
}

func (HybridStrategy) PreferredBatchSize() int { return poolBatchSize }

func (s HybridStrategy) GradeBatch(ctx context.Context, q GradingQuery, docs []GradingDoc) ([]GradedDoc, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	bm25Norms := s.bm25.normScores(q, docs)
	sims, err := s.vector.similarities(ctx, q, docs)
	if err != nil {
		return nil, fmt.Errorf("hybrid vector component: %w", err)
	}

	out := make([]GradedDoc, len(docs))
	for i, d := range docs {
		cos := sims[i]
		if cos < 0 {
			cos = 0 // cosine can be negative; clamp so it doesn't cancel BM25
		}
		combined := hybridBM25Weight*bm25Norms[i] + hybridVectorWeight*cos
		out[i] = GradedDoc{DocID: d.ID, Grade: gradeFromNorm(combined)}
	}
	return out, nil
}

func (s HybridStrategy) Grade(ctx context.Context, q GradingQuery, doc GradingDoc) (int, error) {
	res, err := s.GradeBatch(ctx, q, []GradingDoc{doc})
	if err != nil {
		return GradeUnjudged, err
	}
	return res[0].Grade, nil
}
