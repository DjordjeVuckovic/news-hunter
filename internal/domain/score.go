package domain

const ScoreDecimalPlaces = 4

func NormalizeScore(rawScore *float64) float64 {
	maxScore := 1.0
	if rawScore != nil && *rawScore > 0 {
		maxScore = *rawScore
	}
	return maxScore
}
