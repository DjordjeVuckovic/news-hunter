package report

import (
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/bench/runner"
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/spec"
)

const Version = "1.0.0"

type GenerateOptions struct {
	Spec   *spec.BenchSpec
	Corpus CorpusInfo
}

func Generate(br *runner.BenchmarkResult, opts *GenerateOptions) *Report {
	r := &Report{
		Meta: BenchMeta{
			Version:     Version,
			Timestamp:   time.Now().UTC(),
			Engines:     make(map[string]EngineInfo),
			Environment: NewEnvironmentInfo(),
		},
		Config: ReportConfig{
			KValues:            br.Config.KValues,
			RelevanceThreshold: br.Config.RelevanceThreshold,
		},
	}

	if opts != nil {
		if opts.Spec != nil {
			for name, eng := range opts.Spec.Engines {
				r.Meta.Engines[name] = EngineInfo{
					Type:       eng.Type,
					Connection: maskConnection(eng.Connection),
				}
			}
		}
		r.Meta.Corpus = opts.Corpus
	}

	for _, jr := range br.Jobs {
		jobReport := generateJobReport(jr, br.Config.KValues)
		r.Jobs = append(r.Jobs, jobReport)
	}

	return r
}

func maskConnection(conn string) string {
	if len(conn) > 50 {
		return conn[:20] + "..." + conn[len(conn)-20:]
	}
	return conn
}

func generateJobReport(jr *runner.JobResult, kValues []int) JobReport {
	report := JobReport{
		JobName: jr.JobName,
	}

	for _, qID := range jr.QueryOrder {
		engineResults := jr.Results[qID]
		for _, engName := range jr.EngineNames {
			qr, ok := engineResults[engName]
			if !ok {
				continue
			}
			entry := Entry{
				QueryID:      qr.QueryID,
				JobName:      qr.JobName,
				EngineName:   qr.EngineName,
				Judged:       qr.Scores.Judged,
				NDCG:         qr.Scores.NDCG,
				Precision:    qr.Scores.Precision,
				Recall:       qr.Scores.Recall,
				F1:           qr.Scores.F1,
				AP:           qr.Scores.AP,
				RR:           qr.Scores.RR,
				Bpref:        qr.Scores.Bpref,
				TotalMatches: qr.TotalMatches,
				Latency:      fromRunnerLatencyStats(qr.Latency),
			}
			if qr.Error != nil {
				entry.Error = qr.Error.Error()
			}
			report.PerQuery = append(report.PerQuery, entry)
		}
	}

	report.Aggregated = aggregate(jr, kValues)
	return report
}

func aggregate(jr *runner.JobResult, kValues []int) []AggregatedEntry {
	entries := make([]AggregatedEntry, 0, len(jr.EngineNames))

	for _, engName := range jr.EngineNames {
		agg := AggregatedEntry{
			EngineName: engName,
			NDCG:       make(map[int]float64, len(kValues)),
			Precision:  make(map[int]float64, len(kValues)),
			Recall:     make(map[int]float64, len(kValues)),
			F1:         make(map[int]float64, len(kValues)),
		}

		var allStats []runner.LatencyStats

		for _, qID := range jr.QueryOrder {
			qr, ok := jr.Results[qID][engName]
			if !ok {
				continue
			}
			agg.QueryCount++

			if qr.Error != nil {
				agg.ErrorCount++
				continue
			}

			allStats = append(allStats, qr.Latency)

			if !qr.Scores.Judged {
				continue
			}

			agg.JudgedCount++
			agg.MAP += qr.Scores.AP
			agg.MRR += qr.Scores.RR
			agg.MBpref += qr.Scores.Bpref

			for _, k := range kValues {
				agg.NDCG[k] += qr.Scores.NDCG[k]
				agg.Precision[k] += qr.Scores.Precision[k]
				agg.Recall[k] += qr.Scores.Recall[k]
				agg.F1[k] += qr.Scores.F1[k]
			}
		}

		if len(allStats) > 0 {
			aggregatedStats := runner.AggregateLatencyStats(allStats)
			agg.Latency = fromRunnerLatencyStats(aggregatedStats)
		}

		if agg.JudgedCount > 0 {
			n := float64(agg.JudgedCount)
			agg.MAP /= n
			agg.MRR /= n
			agg.MBpref /= n

			for _, k := range kValues {
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
