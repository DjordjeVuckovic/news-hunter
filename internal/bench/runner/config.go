package runner

var DefaultKValues = []int{3, 5, 10}

const (
	DefaultMaxK               = 10
	DefaultRelevanceThreshold = 1
)

type Config struct {
	KValues            []int
	MaxK               int
	RelevanceThreshold int
}

func DefaultConfig() Config {
	return Config{
		KValues:            DefaultKValues,
		MaxK:               DefaultMaxK,
		RelevanceThreshold: DefaultRelevanceThreshold,
	}
}
