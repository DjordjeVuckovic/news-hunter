package report

import (
	"runtime"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
)

type Report struct {
	Meta   BenchMeta    `json:"meta"`
	Jobs   []JobReport  `json:"jobs"`
	Config ReportConfig `json:"config"`
}

type BenchMeta struct {
	Version     string                `json:"version"`
	Timestamp   time.Time             `json:"timestamp"`
	Engines     map[string]EngineInfo `json:"engines"`
	Corpus      CorpusInfo            `json:"corpus,omitempty"`
	Environment EnvironmentInfo       `json:"environment"`
}

type EngineInfo struct {
	Type       string `json:"type"`
	Connection string `json:"connection"`
	Version    string `json:"version,omitempty"`
}

type CorpusInfo struct {
	Name      string `json:"name,omitempty"`
	DocCount  int64  `json:"doc_count,omitempty"`
	IndexName string `json:"index_name,omitempty"`
}

type EnvironmentInfo struct {
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	NumCPU    int    `json:"num_cpu"`
}

func NewEnvironmentInfo() EnvironmentInfo {
	return EnvironmentInfo{
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
	}
}

type JobReport struct {
	JobName    string
	Aggregated []AggregatedEntry
	PerQuery   []Entry
}

type ReportConfig struct {
	KValues            []int
	RelevanceThreshold int
}

type Entry struct {
	QueryID      string
	JobName      string
	EngineName   string
	NDCG         map[int]float64
	Precision    map[int]float64
	Recall       map[int]float64
	F1           map[int]float64
	AP           float64
	RR           float64
	TotalMatches int64
	Latency      LatencyStats
	Error        string
}

type AggregatedEntry struct {
	EngineName string
	NDCG       map[int]float64
	Precision  map[int]float64
	Recall     map[int]float64
	F1         map[int]float64
	MAP        float64
	MRR        float64
	Latency    LatencyStats
	QueryCount int
	ErrorCount int
}

type LatencyStats struct {
	Min         time.Duration         `json:"min"`
	Max         time.Duration         `json:"max"`
	Mean        time.Duration         `json:"mean"`
	Median      time.Duration         `json:"median"`
	Stddev      time.Duration         `json:"stddev"`
	Percentiles map[int]time.Duration `json:"percentiles"`
	SampleCount int                   `json:"sample_count"`
}

func fromRunnerLatencyStats(s runner.LatencyStats) LatencyStats {
	return LatencyStats{
		Min:         s.Min,
		Max:         s.Max,
		Mean:        s.Mean,
		Median:      s.Median,
		Stddev:      s.Stddev,
		Percentiles: s.Percentiles,
		SampleCount: s.SampleCount,
	}
}

func (s LatencyStats) P50() time.Duration { return s.Percentiles[50] }
func (s LatencyStats) P75() time.Duration { return s.Percentiles[75] }
func (s LatencyStats) P90() time.Duration { return s.Percentiles[90] }
func (s LatencyStats) P95() time.Duration { return s.Percentiles[95] }
func (s LatencyStats) P99() time.Duration { return s.Percentiles[99] }
