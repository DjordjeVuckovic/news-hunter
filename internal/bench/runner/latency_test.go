package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComputeLatencyStats_Empty(t *testing.T) {
	stats := ComputeLatencyStats(nil)
	assert.Zero(t, stats.Min)
	assert.Zero(t, stats.Max)
	assert.Zero(t, stats.Mean)
	assert.Zero(t, stats.Median)
	assert.Zero(t, stats.SampleCount)
	assert.True(t, stats.IsZero())
}

func TestComputeLatencyStats_SingleValue(t *testing.T) {
	durations := []time.Duration{10 * time.Millisecond}
	stats := ComputeLatencyStats(durations)

	assert.Equal(t, 10*time.Millisecond, stats.Min)
	assert.Equal(t, 10*time.Millisecond, stats.Max)
	assert.Equal(t, 10*time.Millisecond, stats.Mean)
	assert.Equal(t, 10*time.Millisecond, stats.Median)
	assert.Equal(t, 1, stats.SampleCount)
	assert.Zero(t, stats.Stddev)
	assert.False(t, stats.IsZero())
}

func TestComputeLatencyStats_MultipleValues(t *testing.T) {
	durations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}
	stats := ComputeLatencyStats(durations)

	assert.Equal(t, 10*time.Millisecond, stats.Min)
	assert.Equal(t, 50*time.Millisecond, stats.Max)
	assert.Equal(t, 30*time.Millisecond, stats.Mean)
	assert.Equal(t, 30*time.Millisecond, stats.Median)
	assert.Equal(t, 5, stats.SampleCount)
}

func TestComputeLatencyStats_EvenCount(t *testing.T) {
	durations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
	}
	stats := ComputeLatencyStats(durations)

	assert.Equal(t, 10*time.Millisecond, stats.Min)
	assert.Equal(t, 40*time.Millisecond, stats.Max)
	assert.Equal(t, 25*time.Millisecond, stats.Mean)
	assert.Equal(t, 4, stats.SampleCount)
}

func TestComputeLatencyStats_Percentiles(t *testing.T) {
	durations := make([]time.Duration, 100)
	for i := range durations {
		durations[i] = time.Duration(i+1) * time.Millisecond
	}
	stats := ComputeLatencyStats(durations)

	assert.Equal(t, 1*time.Millisecond, stats.Min)
	assert.Equal(t, 100*time.Millisecond, stats.Max)
	assert.Equal(t, 100, stats.SampleCount)

	assert.InDelta(t, float64(50*time.Millisecond), float64(stats.P50()), float64(1*time.Millisecond))
	assert.InDelta(t, float64(75*time.Millisecond), float64(stats.P75()), float64(1*time.Millisecond))
	assert.InDelta(t, float64(90*time.Millisecond), float64(stats.P90()), float64(1*time.Millisecond))
	assert.InDelta(t, float64(95*time.Millisecond), float64(stats.P95()), float64(1*time.Millisecond))
	assert.InDelta(t, float64(99*time.Millisecond), float64(stats.P99()), float64(1*time.Millisecond))
}

func TestComputeLatencyStats_Stddev(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}
	stats := ComputeLatencyStats(durations)
	assert.Zero(t, stats.Stddev)

	durations2 := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
	}
	stats2 := ComputeLatencyStats(durations2)
	assert.Greater(t, stats2.Stddev, time.Duration(0))
}

func TestComputeLatencyStats_Unsorted(t *testing.T) {
	durations := []time.Duration{
		50 * time.Millisecond,
		10 * time.Millisecond,
		30 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
	}
	stats := ComputeLatencyStats(durations)

	assert.Equal(t, 10*time.Millisecond, stats.Min)
	assert.Equal(t, 50*time.Millisecond, stats.Max)
	assert.Equal(t, 30*time.Millisecond, stats.Median)
}

func TestAggregateLatencyStats(t *testing.T) {
	stats1 := ComputeLatencyStats([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond})
	stats2 := ComputeLatencyStats([]time.Duration{30 * time.Millisecond, 40 * time.Millisecond})

	agg := AggregateLatencyStats([]LatencyStats{stats1, stats2})

	assert.Equal(t, 10*time.Millisecond, agg.Min)
	assert.Equal(t, 40*time.Millisecond, agg.Max)
	assert.Equal(t, 4, agg.SampleCount)
	assert.Equal(t, 25*time.Millisecond, agg.Mean)
}

func TestAggregateLatencyStats_Empty(t *testing.T) {
	agg := AggregateLatencyStats(nil)
	assert.True(t, agg.IsZero())
}

func TestPercentile_EdgeCases(t *testing.T) {
	sorted := []time.Duration{10 * time.Millisecond}
	assert.Equal(t, 10*time.Millisecond, percentile(sorted, 0))
	assert.Equal(t, 10*time.Millisecond, percentile(sorted, 50))
	assert.Equal(t, 10*time.Millisecond, percentile(sorted, 100))

	empty := []time.Duration{}
	assert.Zero(t, percentile(empty, 50))
}
