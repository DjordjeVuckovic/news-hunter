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
