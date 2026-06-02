package judgment

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestBM25GradeBatch_RanksByRelevance(t *testing.T) {
	q := GradingQuery{ID: "q1", Description: "climate change policy"}
	strong := GradingDoc{ID: uuid.New(), Title: "Climate change policy reshapes energy", Description: "New climate policy targets emissions"}
	weak := GradingDoc{ID: uuid.New(), Title: "Climate mentioned once", Description: "A short note"}
	none := GradingDoc{ID: uuid.New(), Title: "Local football results", Description: "Match report"}

	got, err := BM25Strategy{}.GradeBatch(context.Background(), q, []GradingDoc{strong, weak, none})
	if err != nil {
		t.Fatalf("GradeBatch: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d grades, want 3", len(got))
	}
	byID := map[uuid.UUID]int{}
	for _, g := range got {
		byID[g.DocID] = g.Grade
	}
	if byID[strong.ID] <= byID[weak.ID] {
		t.Errorf("strong (%d) should outrank weak (%d)", byID[strong.ID], byID[weak.ID])
	}
	if byID[none.ID] != GradeNotRelev {
		t.Errorf("non-matching doc grade = %d, want %d", byID[none.ID], GradeNotRelev)
	}
}

func TestBM25GradeBatch_EmptyQueryDescription(t *testing.T) {
	q := GradingQuery{ID: "q1", Description: "the a of"} // all stopwords/short
	doc := GradingDoc{ID: uuid.New(), Title: "anything"}
	got, err := BM25Strategy{}.GradeBatch(context.Background(), q, []GradingDoc{doc})
	if err != nil {
		t.Fatalf("GradeBatch: %v", err)
	}
	if got[0].Grade != GradeNotRelev {
		t.Errorf("grade = %d, want %d for empty query", got[0].Grade, GradeNotRelev)
	}
}

func TestGradeFromNorm(t *testing.T) {
	cases := []struct {
		norm float64
		want int
	}{
		{1.0, GradeHighly},
		{0.66, GradeHighly},
		{0.5, GradeRelevant},
		{0.40, GradeRelevant},
		{0.2, GradeMarginally},
		{0.15, GradeMarginally},
		{0.1, GradeNotRelev},
		{0.0, GradeNotRelev},
	}
	for _, c := range cases {
		if got := gradeFromNorm(c.norm); got != c.want {
			t.Errorf("gradeFromNorm(%v) = %d, want %d", c.norm, got, c.want)
		}
	}
}
