package metrics

import "github.com/google/uuid"

// AveragePrecision computes the mean of precision values at each relevant rank position.
func AveragePrecision(ranked []uuid.UUID, judgments map[uuid.UUID]int, relevanceThreshold int) float64 {
	if len(ranked) == 0 {
		return 0
	}

	totalRelevant := countRelevant(judgments, relevanceThreshold)
	if totalRelevant == 0 {
		return 0
	}

	var sumPrecision float64
	var relevantSeen int

	for i, docID := range ranked {
		if judgments[docID] >= relevanceThreshold {
			relevantSeen++
			sumPrecision += float64(relevantSeen) / float64(i+1)
		}
	}

	return sumPrecision / float64(totalRelevant)
}

// ReciprocalRank returns 1/rank of the first relevant document.
func ReciprocalRank(ranked []uuid.UUID, judgments map[uuid.UUID]int, relevanceThreshold int) float64 {
	for i, docID := range ranked {
		if judgments[docID] >= relevanceThreshold {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}
