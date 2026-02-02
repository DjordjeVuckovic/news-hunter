package suite

import (
	"testing"

	"github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("valid suite", func(t *testing.T) {
		yaml := `
name: test
version: "1.0"
queries:
  - id: q1
    description: basic query string
    kind: query_string
    query_string:
      query: "climate change"
    judgments: []
`
		s, err := Parse([]byte(yaml))
		require.NoError(t, err)
		assert.Equal(t, "test", s.Name)
		assert.Len(t, s.Queries, 1)
		assert.Equal(t, "q1", s.Queries[0].ID)
		assert.Equal(t, query.QueryStringType, s.Queries[0].Kind)
	})

	t.Run("empty queries", func(t *testing.T) {
		yaml := `
name: test
queries: []
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no queries")
	})

	t.Run("missing id", func(t *testing.T) {
		yaml := `
name: test
queries:
  - kind: query_string
    query_string:
      query: test
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no id")
	})

	t.Run("missing kind", func(t *testing.T) {
		yaml := `
name: test
queries:
  - id: q1
    query_string:
      query: test
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no kind")
	})
}

func TestParse_AllQueryKinds(t *testing.T) {
	docID := uuid.New()

	yaml := `
name: all kinds
version: "1.0"
queries:
  - id: qs1
    description: query string
    kind: query_string
    query_string:
      query: "test query"
      language: english
      operator: OR
    judgments:
      - doc_id: ` + docID.String() + `
        relevance: 3

  - id: m1
    description: match
    kind: match
    match:
      query: "test"
      field: title
      operator: AND

  - id: mm1
    description: multi match
    kind: multi_match
    multi_match:
      query: "test"
      fields: ["title", "content"]

  - id: p1
    description: phrase
    kind: phrase
    phrase:
      query: "exact phrase"
      fields: ["title"]
      slop: 2

  - id: b1
    description: boolean
    kind: boolean
    boolean:
      expression: "climate AND change"
`
	s, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Len(t, s.Queries, 5)

	// Verify first query has judgment
	assert.Len(t, s.Queries[0].Judgments, 1)
	assert.Equal(t, docID, s.Queries[0].Judgments[0].DocID)
	assert.Equal(t, 3, s.Queries[0].Judgments[0].Relevance)
}

func TestToDomainQuery(t *testing.T) {
	t.Run("query_string", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "qs1",
			Kind: query.QueryStringType,
			QueryString: &QueryStringSpec{
				Query:    "climate change",
				Language: "english",
				Operator: "OR",
			},
		}
		base, err := ToDomainQuery(bq)
		require.NoError(t, err)
		assert.NotNil(t, base.QueryString)
		assert.Equal(t, "climate change", base.QueryString.Query)
	})

	t.Run("match", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "m1",
			Kind: query.MatchType,
			Match: &MatchSpec{
				Query:    "climate",
				Field:    "title",
				Operator: "AND",
			},
		}
		base, err := ToDomainQuery(bq)
		require.NoError(t, err)
		assert.NotNil(t, base.Match)
		assert.Equal(t, "title", base.Match.Field)
	})

	t.Run("multi_match", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "mm1",
			Kind: query.MultiMatchType,
			MultiMatch: &MultiMatchSpec{
				Query:  "climate",
				Fields: []string{"title", "content"},
			},
		}
		base, err := ToDomainQuery(bq)
		require.NoError(t, err)
		assert.NotNil(t, base.MultiMatch)
		assert.Len(t, base.MultiMatch.Fields, 2)
	})

	t.Run("phrase", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "p1",
			Kind: query.PhraseType,
			Phrase: &PhraseSpec{
				Query:  "climate change",
				Fields: []string{"title"},
				Slop:   2,
			},
		}
		base, err := ToDomainQuery(bq)
		require.NoError(t, err)
		assert.NotNil(t, base.Phrase)
		assert.Equal(t, 2, base.Phrase.Slop)
	})

	t.Run("boolean", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "b1",
			Kind: query.BooleanType,
			Boolean: &BooleanSpec{
				Expression: "climate AND change",
			},
		}
		base, err := ToDomainQuery(bq)
		require.NoError(t, err)
		assert.NotNil(t, base.Boolean)
		assert.Equal(t, "climate AND change", base.Boolean.Expression)
	})

	t.Run("nil spec", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "bad",
			Kind: query.QueryStringType,
		}
		_, err := ToDomainQuery(bq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("unsupported kind", func(t *testing.T) {
		bq := &BenchmarkQuery{
			ID:   "bad",
			Kind: "unknown",
		}
		_, err := ToDomainQuery(bq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported kind")
	})
}

func TestBenchmarkQuery_JudgmentMap(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	bq := BenchmarkQuery{
		Judgments: []RelevanceJudgment{
			{DocID: id1, Relevance: 3},
			{DocID: id2, Relevance: 1},
		},
	}

	m := bq.JudgmentMap()
	assert.Equal(t, 3, m[id1])
	assert.Equal(t, 1, m[id2])
}
