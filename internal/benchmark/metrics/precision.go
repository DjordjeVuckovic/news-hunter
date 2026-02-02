package metrics

import "github.com/google/uuid"

// PrecisionAtK computes the fraction of top-K results that are relevant.
// A document is relevant if its judgment >= relevanceThreshold.
func PrecisionAtK(ranked []uuid.UUID, judgments map[uuid.UUID]int, k int, relevanceThreshold int) float64 {
	if k <= 0 || len(ranked) == 0 {
		return 0
	}

	n := min(k, len(ranked))
	var relevant int

	for i := 0; i < n; i++ {
		if judgments[ranked[i]] >= relevanceThreshold {
			relevant++
		}
	}

	return float64(relevant) / float64(k)
}

// RecallAtK computes the fraction of all relevant documents found in top-K.
func RecallAtK(ranked []uuid.UUID, judgments map[uuid.UUID]int, k int, relevanceThreshold int) float64 {
	if k <= 0 || len(ranked) == 0 {
		return 0
	}

	totalRelevant := countRelevant(judgments, relevanceThreshold)
	if totalRelevant == 0 {
		return 0
	}

	n := min(k, len(ranked))
	var found int

	for i := 0; i < n; i++ {
		if judgments[ranked[i]] >= relevanceThreshold {
			found++
		}
	}

	return float64(found) / float64(totalRelevant)
}

// F1AtK computes the harmonic mean of P@K and R@K.
func F1AtK(ranked []uuid.UUID, judgments map[uuid.UUID]int, k int, relevanceThreshold int) float64 {
	p := PrecisionAtK(ranked, judgments, k, relevanceThreshold)
	r := RecallAtK(ranked, judgments, k, relevanceThreshold)

	if p+r == 0 {
		return 0
	}

	return 2 * p * r / (p + r)
}

func countRelevant(judgments map[uuid.UUID]int, threshold int) int {
	var count int
	for _, rel := range judgments {
		if rel >= threshold {
			count++
		}
	}
	return count
}
