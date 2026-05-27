package metrics

import (
	"math"
	"testing"
)

func TestWilcoxon(t *testing.T) {
	tests := []struct {
		name      string
		a, b      []float64
		wantNil   bool
		wantStars string
		wantP     float64 // approx upper bound
	}{
		{
			name:    "too few pairs returns nil",
			a:       []float64{0.5, 0.6, 0.7},
			b:       []float64{0.5, 0.6, 0.7},
			wantNil: true,
		},
		{
			name:    "all tied returns nil",
			a:       []float64{0.5, 0.5, 0.5, 0.5, 0.5},
			b:       []float64{0.5, 0.5, 0.5, 0.5, 0.5},
			wantNil: true,
		},
		{
			name:      "no significant difference",
			a:         []float64{0.50, 0.55, 0.48, 0.52, 0.53, 0.49, 0.51, 0.54, 0.47, 0.56},
			b:         []float64{0.51, 0.54, 0.50, 0.51, 0.52, 0.51, 0.50, 0.53, 0.49, 0.55},
			wantStars: "",
		},
		{
			name:      "engine A clearly dominates — significant",
			a:         []float64{0.90, 0.85, 0.88, 0.92, 0.87, 0.91, 0.86, 0.89, 0.93, 0.84},
			b:         []float64{0.30, 0.25, 0.28, 0.32, 0.27, 0.31, 0.26, 0.29, 0.33, 0.24},
			wantStars: "**",
			wantP:     0.01,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Wilcoxon(tc.a, tc.b)
			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if got.Stars != tc.wantStars {
				t.Errorf("stars: got %q, want %q (p=%.4f)", got.Stars, tc.wantStars, got.P)
			}
			if tc.wantP > 0 && got.P > tc.wantP {
				t.Errorf("p-value %.4f exceeds expected upper bound %.4f", got.P, tc.wantP)
			}
			if math.IsNaN(got.P) || math.IsInf(got.P, 0) {
				t.Errorf("p-value is not finite: %v", got.P)
			}
		})
	}
}
