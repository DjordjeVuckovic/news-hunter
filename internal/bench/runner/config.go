package runner

var DefaultKValues = []int{3, 5, 10}

const (
	DefaultMaxK               = 10
	DefaultRelevanceThreshold = 1
	DefaultWarmupRuns         = 0
	DefaultRuns               = 1
)

type Config struct {
	KValues            []int
	MaxK               int
	RelevanceThreshold int
	WarmupRuns         int
	Runs               int
	// Judgments[queryID][docID]grade — pre-loaded by the CLI from the
	// resolved annotations file. When nil, queries are scored without
	// relevance grades and the report flags them as unjudged.
	Judgments map[string]map[string]int
}

func DefaultConfig() Config {
	return Config{
		KValues:            DefaultKValues,
		MaxK:               DefaultMaxK,
		RelevanceThreshold: DefaultRelevanceThreshold,
		WarmupRuns:         DefaultWarmupRuns,
		Runs:               DefaultRuns,
	}
}
