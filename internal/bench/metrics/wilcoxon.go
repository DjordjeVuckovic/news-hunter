package metrics

import (
	"math"
	"sort"
)

// PairwiseResult is the outcome of a two-tailed Wilcoxon signed-rank test
// comparing per-query scores of two engines on one metric.
type PairwiseResult struct {
	W     float64 // test statistic (smaller of W+ and W-)
	P     float64 // two-tailed p-value (normal approximation with continuity correction)
	Stars string  // significance marker: "**" p<0.01, "*" p<0.05, "" not significant
}

// Wilcoxon performs the two-tailed Wilcoxon signed-rank test on paired scores.
// a[i] and b[i] are per-query scores for the same query under two engines.
// Uses the normal approximation with continuity correction — adequate for n≥6.
// Returns nil when fewer than 4 non-tied pairs exist (insufficient data).
func Wilcoxon(a, b []float64) *PairwiseResult {
	if len(a) != len(b) || len(a) == 0 {
		return nil
	}

	type observation struct {
		absD float64
		sign float64
	}

	var obs []observation
	for i := range a {
		d := a[i] - b[i]
		if d == 0 {
			continue // tied differences are excluded from the test
		}
		s := 1.0
		if d < 0 {
			s = -1.0
		}
		obs = append(obs, observation{math.Abs(d), s})
	}
	n := len(obs)
	if n < 4 {
		return nil
	}

	// Sort by absolute difference ascending (smallest rank = 1).
	sort.Slice(obs, func(i, j int) bool { return obs[i].absD < obs[j].absD })

	// Assign average ranks for tied absolute differences; accumulate tie
	// correction term for the variance formula.
	ranks := make([]float64, n)
	tieCorr := 0.0
	for i := 0; i < n; {
		j := i
		for j < n && obs[j].absD == obs[i].absD {
			j++
		}
		avg := float64(i+1+j) / 2.0 // 1-indexed average
		t := float64(j - i)
		if t > 1 {
			tieCorr += t*t*t - t
		}
		for k := i; k < j; k++ {
			ranks[k] = avg
		}
		i = j
	}

	wPlus, wMinus := 0.0, 0.0
	for i, o := range obs {
		if o.sign > 0 {
			wPlus += ranks[i]
		} else {
			wMinus += ranks[i]
		}
	}
	W := math.Min(wPlus, wMinus)

	fn := float64(n)
	mu := fn * (fn + 1) / 4
	variance := fn*(fn+1)*(2*fn+1)/24 - tieCorr/48
	if variance <= 0 {
		return nil
	}
	// Continuity correction: shift W toward the mean by 0.5.
	z := (W - mu + 0.5) / math.Sqrt(variance)
	p := 2 * normCDF(-math.Abs(z))

	stars := ""
	switch {
	case p < 0.01:
		stars = "**"
	case p < 0.05:
		stars = "*"
	}
	return &PairwiseResult{W: W, P: p, Stars: stars}
}

// normCDF returns P(Z ≤ x) for the standard normal distribution.
func normCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}
