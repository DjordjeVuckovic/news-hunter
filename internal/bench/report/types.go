package report

import "time"

type Report struct {
	Aggregated []AggregatedEntry
	PerQuery   []Entry
	Config     ReportConfig
}

type ReportConfig struct {
	KValues            []int
	RelevanceThreshold int
}

type Entry struct {
	QueryID      string
	EngineName   string
	NDCG         map[int]float64
	Precision    map[int]float64
	Recall       map[int]float64
	F1           map[int]float64
	AP           float64
	RR           float64
	TotalMatches int64
	Latency      time.Duration
	Error        string
}

type AggregatedEntry struct {
	EngineName  string
	NDCG        map[int]float64
	Precision   map[int]float64
	Recall      map[int]float64
	F1          map[int]float64
	MAP         float64
	MRR         float64
	MeanLatency time.Duration
	QueryCount  int
	ErrorCount  int
}
