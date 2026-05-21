package spec

type BenchSpec struct {
	SchemaVersion int               `yaml:"schema_version"`
	ID            string            `yaml:"id"`
	Description   string            `yaml:"description,omitempty"`
	Defaults      Defaults          `yaml:"defaults,omitempty"`
	Engines       map[string]Engine `yaml:"engines"`
	Metrics       MetricsConfig     `yaml:"metrics"`
	Runs          RunsConfig        `yaml:"runs"`
	Jobs          []Job             `yaml:"jobs"`
}

// Defaults supply fallback values that the CLI flags can override. Lets users
// set "this track defaults to lexical judgments and pool depth 100" in one
// place instead of repeating flags.
type Defaults struct {
	PoolDepth int    `yaml:"pool_depth,omitempty"`
	Judgments string `yaml:"judgments,omitempty"` // strategy name OR path
}

type Job struct {
	Name    string   `yaml:"name"`
	Suite   string   `yaml:"suite"`
	Engines []string `yaml:"engines"`
}

type Engine struct {
	Type       string `yaml:"type"`
	Connection string `yaml:"connection"`
	Index      string `yaml:"index,omitempty"`
}

type MetricsConfig struct {
	KValues            []int `yaml:"k_values"`
	MaxK               int   `yaml:"max_k"`
	RelevanceThreshold int   `yaml:"relevance_threshold"`
}

type RunsConfig struct {
	Warmup     int `yaml:"warmup"`
	Iterations int `yaml:"iterations"`
}
