package runner

import (
	"math"
	"sort"
	"time"
)

type LatencyStats struct {
	Min         time.Duration         `json:"min"`
	Max         time.Duration         `json:"max"`
	Mean        time.Duration         `json:"mean"`
	Median      time.Duration         `json:"median"`
	Stddev      time.Duration         `json:"stddev"`
	Percentiles map[int]time.Duration `json:"percentiles"`
	SampleCount int                   `json:"sample_count"`
	Raw         []time.Duration       `json:"-"`
}

var defaultPercentiles = []int{50, 75, 90, 95, 99}

func ComputeLatencyStats(durations []time.Duration) LatencyStats {
	if len(durations) == 0 {
		return LatencyStats{
			Percentiles: make(map[int]time.Duration),
		}
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	stats := LatencyStats{
		Min:         sorted[0],
		Max:         sorted[len(sorted)-1],
		Median:      percentile(sorted, 50),
		Percentiles: make(map[int]time.Duration),
		SampleCount: len(durations),
		Raw:         durations,
	}

	var sum int64
	for _, d := range sorted {
		sum += int64(d)
	}
	stats.Mean = time.Duration(sum / int64(len(sorted)))

	if len(sorted) > 1 {
		var sumSquares float64
		meanNs := float64(stats.Mean.Nanoseconds())
		for _, d := range sorted {
			diff := float64(d.Nanoseconds()) - meanNs
			sumSquares += diff * diff
		}
		variance := sumSquares / float64(len(sorted)-1)
		stats.Stddev = time.Duration(math.Sqrt(variance))
	}

	for _, p := range defaultPercentiles {
		stats.Percentiles[p] = percentile(sorted, p)
	}

	return stats
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(rank)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	weight := rank - float64(lower)
	return time.Duration(float64(sorted[lower])*(1-weight) + float64(sorted[upper])*weight)
}

func AggregateLatencyStats(stats []LatencyStats) LatencyStats {
	if len(stats) == 0 {
		return LatencyStats{Percentiles: make(map[int]time.Duration)}
	}

	var allDurations []time.Duration
	for _, s := range stats {
		allDurations = append(allDurations, s.Raw...)
	}

	return ComputeLatencyStats(allDurations)
}

func (s LatencyStats) P50() time.Duration { return s.Percentiles[50] }
func (s LatencyStats) P75() time.Duration { return s.Percentiles[75] }
func (s LatencyStats) P90() time.Duration { return s.Percentiles[90] }
func (s LatencyStats) P95() time.Duration { return s.Percentiles[95] }
func (s LatencyStats) P99() time.Duration { return s.Percentiles[99] }

func (s LatencyStats) IsZero() bool {
	return s.SampleCount == 0
}
