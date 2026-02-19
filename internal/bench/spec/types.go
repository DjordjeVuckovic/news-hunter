package spec

type BenchSpec struct {
	Jobs    []Job             `yaml:"jobs"`
	Engines map[string]Engine `yaml:"engines"`
	Metrics MetricsConfig     `yaml:"metrics"`
	Runs    RunsConfig        `yaml:"runs"`
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
