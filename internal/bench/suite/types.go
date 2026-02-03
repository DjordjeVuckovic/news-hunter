package suite

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"github.com/google/uuid"
)

type TestSuite struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version"`
	Queries     []BenchmarkQuery `yaml:"queries"`
}

type BenchmarkQuery struct {
	ID          string              `yaml:"id"`
	Description string              `yaml:"description"`
	Kind        query.Kind          `yaml:"kind"`
	QueryString *QueryStringSpec    `yaml:"query_string,omitempty"`
	Match       *MatchSpec          `yaml:"match,omitempty"`
	MultiMatch  *MultiMatchSpec     `yaml:"multi_match,omitempty"`
	Phrase      *PhraseSpec         `yaml:"phrase,omitempty"`
	Boolean     *BooleanSpec        `yaml:"boolean,omitempty"`
	Judgments   []RelevanceJudgment `yaml:"judgments"`
}

type RelevanceJudgment struct {
	DocID     uuid.UUID `yaml:"doc_id"`
	Relevance int       `yaml:"relevance"`
}

type QueryStringSpec struct {
	Query    string `yaml:"query"`
	Language string `yaml:"language,omitempty"`
	Operator string `yaml:"operator,omitempty"`
}

type MatchSpec struct {
	Query     string `yaml:"query"`
	Field     string `yaml:"field"`
	Language  string `yaml:"language,omitempty"`
	Operator  string `yaml:"operator,omitempty"`
	Fuzziness string `yaml:"fuzziness,omitempty"`
}

type MultiMatchSpec struct {
	Query    string   `yaml:"query"`
	Fields   []string `yaml:"fields"`
	Language string   `yaml:"language,omitempty"`
	Operator string   `yaml:"operator,omitempty"`
}

type PhraseSpec struct {
	Query    string   `yaml:"query"`
	Fields   []string `yaml:"fields"`
	Slop     int      `yaml:"slop,omitempty"`
	Language string   `yaml:"language,omitempty"`
}

type BooleanSpec struct {
	Expression string `yaml:"expression"`
	Language   string `yaml:"language,omitempty"`
}

// JudgmentMap converts the judgments slice to a map keyed by doc ID.
func (bq *BenchmarkQuery) JudgmentMap() map[uuid.UUID]int {
	m := make(map[uuid.UUID]int, len(bq.Judgments))
	for _, j := range bq.Judgments {
		m[j.DocID] = j.Relevance
	}
	return m
}
