package judgment

import (
	"context"
	"testing"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/document"
	"github.com/google/uuid"
)

// fakeEmbedder returns canned vectors so the scoring logic can be tested with
// no embedding backend.
type fakeEmbedder struct {
	query []float32
	docs  map[uuid.UUID][]float32
}

func (f fakeEmbedder) EmbedQuery(_ context.Context, _ string) (*embedding.Vec, error) {
	return &embedding.Vec{Embedding: f.query}, nil
}

func (f fakeEmbedder) EmbedDocs(_ context.Context, docs []document.Article) ([]embedding.Vec, error) {
	out := make([]embedding.Vec, len(docs))
	for i, d := range docs {
		out[i] = embedding.Vec{ID: d.ID, Embedding: f.docs[d.ID]}
	}
	return out, nil
}

func TestVectorGradeBatch(t *testing.T) {
	identical := GradingDoc{ID: uuid.New(), Title: "a"}  // cos 1.0  -> Highly
	related := GradingDoc{ID: uuid.New(), Title: "b"}    // cos 0.6  -> Relevant
	orthogonal := GradingDoc{ID: uuid.New(), Title: "c"} // cos 0.0  -> NotRelev

	fe := fakeEmbedder{
		query: []float32{1, 0},
		docs: map[uuid.UUID][]float32{
			identical.ID:  {1, 0},
			related.ID:    {0.6, 0.8},
			orthogonal.ID: {0, 1},
		},
	}
	s := NewVectorStrategyWithEmbedder(fe, "fake-model")

	got, err := s.GradeBatch(context.Background(), GradingQuery{ID: "q"}, []GradingDoc{identical, related, orthogonal})
	if err != nil {
		t.Fatalf("GradeBatch: %v", err)
	}
	byID := map[uuid.UUID]int{}
	for _, g := range got {
		byID[g.DocID] = g.Grade
	}
	if byID[identical.ID] != GradeHighly {
		t.Errorf("identical grade = %d, want %d", byID[identical.ID], GradeHighly)
	}
	if byID[related.ID] != GradeRelevant {
		t.Errorf("related grade = %d, want %d", byID[related.ID], GradeRelevant)
	}
	if byID[orthogonal.ID] != GradeNotRelev {
		t.Errorf("orthogonal grade = %d, want %d", byID[orthogonal.ID], GradeNotRelev)
	}
}

func TestVectorModelID(t *testing.T) {
	s := NewVectorStrategyWithEmbedder(fakeEmbedder{}, "qwen3")
	if got := s.ModelID(); got != "vector:qwen3" {
		t.Errorf("ModelID = %q, want %q", got, "vector:qwen3")
	}
}

func TestCosine(t *testing.T) {
	cases := []struct {
		name string
		a, b []float32
		want float64
	}{
		{"identical", []float32{1, 0}, []float32{1, 0}, 1},
		{"orthogonal", []float32{1, 0}, []float32{0, 1}, 0},
		{"opposite", []float32{1, 0}, []float32{-1, 0}, -1},
		{"len mismatch", []float32{1, 0}, []float32{1}, 0},
		{"empty", nil, []float32{1}, 0},
		{"zero vector", []float32{0, 0}, []float32{1, 0}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := cosine(c.a, c.b); got != c.want {
				t.Errorf("cosine = %v, want %v", got, c.want)
			}
		})
	}
}

func TestEmbedderFromOptions_RequiresEndpoint(t *testing.T) {
	if _, _, err := embedderFromOptions(StrategyOptions{}); err == nil {
		t.Fatal("expected error when EmbeddingBaseURL is empty")
	}
}
