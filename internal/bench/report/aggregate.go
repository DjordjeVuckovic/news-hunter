package report

import (
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
)

func Generate(br *runner.BenchmarkResult) *Report {
	r := &Report{
		Config: ReportConfig{
			KValues:            br.Config.KValues,
			RelevanceThreshold: br.Config.RelevanceThreshold,
		},
	}

	for _, qID := range br.QueryOrder {
		engineResults := br.Results[qID]
		for _, engName := range br.EngineNames {
			qr := engineResults[engName]
			entry := Entry{
				QueryID:      qr.QueryID,
				EngineName:   qr.EngineName,
				NDCG:         qr.Scores.NDCG,
				Precision:    qr.Scores.Precision,
				Recall:       qr.Scores.Recall,
				F1:           qr.Scores.F1,
				AP:           qr.Scores.AP,
				RR:           qr.Scores.RR,
				TotalMatches: qr.TotalMatches,
				Latency:      qr.Latency,
			}
			if qr.Error != nil {
				entry.Error = qr.Error.Error()
			}
			r.PerQuery = append(r.PerQuery, entry)
		}
	}

	r.Aggregated = aggregate(br)

	return r
}

func aggregate(br *runner.BenchmarkResult) []AggregatedEntry {
	entries := make([]AggregatedEntry, 0, len(br.EngineNames))

	for _, engName := range br.EngineNames {
		agg := AggregatedEntry{
			EngineName: engName,
			NDCG:       make(map[int]float64, len(br.Config.KValues)),
			Precision:  make(map[int]float64, len(br.Config.KValues)),
			Recall:     make(map[int]float64, len(br.Config.KValues)),
			F1:         make(map[int]float64, len(br.Config.KValues)),
		}

		var totalLatency time.Duration
		counted := 0

		for _, qID := range br.QueryOrder {
			qr := br.Results[qID][engName]
			agg.QueryCount++

			if qr.Error != nil {
				agg.ErrorCount++
				continue
			}

			counted++
			agg.MAP += qr.Scores.AP
			agg.MRR += qr.Scores.RR
			totalLatency += qr.Latency

			for _, k := range br.Config.KValues {
				agg.NDCG[k] += qr.Scores.NDCG[k]
				agg.Precision[k] += qr.Scores.Precision[k]
				agg.Recall[k] += qr.Scores.Recall[k]
				agg.F1[k] += qr.Scores.F1[k]
			}
		}

		if counted > 0 {
			n := float64(counted)
			agg.MAP /= n
			agg.MRR /= n
			agg.MeanLatency = totalLatency / time.Duration(counted)

			for _, k := range br.Config.KValues {
				agg.NDCG[k] /= n
				agg.Precision[k] /= n
				agg.Recall[k] /= n
				agg.F1[k] /= n
			}
		}

		entries = append(entries, agg)
	}

	return entries
}
