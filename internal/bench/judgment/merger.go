package judgment

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/suite"
)

func MergeIntoSuite(jf *JudgmentFile, s *suite.TestSuite) *suite.TestSuite {
	judgeMap := make(map[string][]GradedDoc, len(jf.Queries))
	for _, entry := range jf.Queries {
		judgeMap[entry.QueryID] = entry.Docs
	}

	merged := *s
	merged.Queries = make([]suite.Query, len(s.Queries))
	copy(merged.Queries, s.Queries)

	for i, q := range merged.Queries {
		if docs, ok := judgeMap[q.ID]; ok {
			judgments := make([]suite.RelevanceJudgment, 0, len(docs))
			for _, d := range docs {
				if d.Grade >= 0 {
					judgments = append(judgments, suite.RelevanceJudgment{
						DocID:     d.DocID,
						Relevance: d.Grade,
					})
				}
			}
			merged.Queries[i].Judgments = judgments
		}
	}

	return &merged
}
