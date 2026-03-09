package metrics

import "github.com/google/uuid"

// BinaryPreference measures ranking quality using only judged documents.
// Unlike AP/NDCG, it ignores unjudged documents entirely, making it robust
// when judgment sets are incomplete (typical in pooling workflows).
//
// For each judged relevant document in the ranking, its contribution is:
//
//	1 - min(nonRelevantAbove, R) / R
//
// where R = total number of judged relevant documents.
// Non-relevant here means explicitly judged as non-relevant (in judgments, rel < threshold).
// Unjudged documents (absent from judgments) are skipped.
func BinaryPreference(ranked []uuid.UUID, judgments map[uuid.UUID]int, relevanceThreshold int) float64 {
	if len(judgments) == 0 {
		return 0
	}

	R := countRelevant(judgments, relevanceThreshold)
	if R == 0 {
		return 0
	}

	var sum float64
	var nonRelAbove int

	for _, docID := range ranked {
		rel, judged := judgments[docID]
		if !judged {
			continue
		}
		if rel >= relevanceThreshold {
			penalty := float64(min(nonRelAbove, R)) / float64(R)
			sum += 1 - penalty
		} else {
			nonRelAbove++
		}
	}

	return sum / float64(R)
}
