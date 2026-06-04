package pg

import (
	"math"
	"testing"
)

func TestRRFScore(t *testing.T) {
	const k = 60

	tests := []struct {
		name    string
		k       int
		lexRank int
		vecRank int
		want    float64
	}{
		{
			name:    "present in both lists at rank 1",
			k:       k,
			lexRank: 1,
			vecRank: 1,
			want:    1.0/float64(k+1) + 1.0/float64(k+1),
		},
		{
			name:    "lexical only",
			k:       k,
			lexRank: 3,
			vecRank: 0,
			want:    1.0 / float64(k+3),
		},
		{
			name:    "vector only",
			k:       k,
			lexRank: 0,
			vecRank: 5,
			want:    1.0 / float64(k+5),
		},
		{
			name:    "absent from both",
			k:       k,
			lexRank: 0,
			vecRank: 0,
			want:    0.0,
		},
		{
			name:    "different k flattens contribution",
			k:       1,
			lexRank: 1,
			vecRank: 2,
			want:    1.0/float64(1+1) + 1.0/float64(1+2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rrfScore(tt.k, tt.lexRank, tt.vecRank)
			if math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("rrfScore(%d, %d, %d) = %v, want %v", tt.k, tt.lexRank, tt.vecRank, got, tt.want)
			}
		})
	}
}

func TestRRFScoreRankOrdering(t *testing.T) {
	const k = 60
	// A document ranked 1 in both legs must outscore one ranked lower in both.
	high := rrfScore(k, 1, 1)
	low := rrfScore(k, 10, 10)
	if high <= low {
		t.Fatalf("expected higher-ranked document to score higher: high=%v low=%v", high, low)
	}

	// A document present in both legs must outscore one present in only one leg
	// at the same rank.
	both := rrfScore(k, 2, 2)
	single := rrfScore(k, 2, 0)
	if both <= single {
		t.Fatalf("expected dual-leg document to score higher: both=%v single=%v", both, single)
	}
}
