package judgment

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBatchGradingPrompt(t *testing.T) {
	q := GradingQuery{ID: "q1", Description: "climate change"}
	d1 := GradingDoc{ID: uuid.New(), Title: "Doc one"}
	d2 := GradingDoc{ID: uuid.New(), Title: "Doc two", Description: "desc", Content: "content snippet"}

	out := BuildBatchGradingPrompt(q, []GradingDoc{d1, d2})

	assert.Contains(t, out, "Query: climate change")
	assert.Contains(t, out, "Grade each of the 2 candidate articles")
	assert.Contains(t, out, "[1] doc_id: "+d1.ID.String())
	assert.Contains(t, out, "[2] doc_id: "+d2.ID.String())
	assert.Contains(t, out, "exactly 2 entries")
	assert.Contains(t, out, "No prose, no markdown", "must explicitly forbid markdown wrapping")
}

func TestParseBatchGradeJSON(t *testing.T) {
	d1 := GradingDoc{ID: uuid.New(), Title: "a"}
	d2 := GradingDoc{ID: uuid.New(), Title: "b"}
	d3 := GradingDoc{ID: uuid.New(), Title: "c"}
	expected := []GradingDoc{d1, d2, d3}

	t.Run("complete response", func(t *testing.T) {
		raw := `[
		  {"doc_id":"` + d1.ID.String() + `","grade":2},
		  {"doc_id":"` + d2.ID.String() + `","grade":3},
		  {"doc_id":"` + d3.ID.String() + `","grade":0}
		]`
		parsed, missing, err := ParseBatchGradeJSON(raw, expected)
		require.NoError(t, err)
		assert.Len(t, parsed, 3)
		assert.Empty(t, missing)
	})

	t.Run("response wrapped in prose / fences is tolerated", func(t *testing.T) {
		raw := "Here are the grades:\n```json\n[" +
			`{"doc_id":"` + d1.ID.String() + `","grade":1}` +
			"]\n```\n"
		parsed, missing, err := ParseBatchGradeJSON(raw, expected)
		require.NoError(t, err)
		assert.Len(t, parsed, 1)
		assert.Len(t, missing, 2, "the LLM only returned d1; d2 and d3 should be flagged missing")
	})

	t.Run("preserves input order", func(t *testing.T) {
		raw := `[
		  {"doc_id":"` + d3.ID.String() + `","grade":3},
		  {"doc_id":"` + d1.ID.String() + `","grade":1},
		  {"doc_id":"` + d2.ID.String() + `","grade":2}
		]`
		parsed, _, err := ParseBatchGradeJSON(raw, expected)
		require.NoError(t, err)
		require.Len(t, parsed, 3)
		assert.Equal(t, d1.ID, parsed[0].DocID, "should reorder to input sequence")
		assert.Equal(t, d2.ID, parsed[1].DocID)
		assert.Equal(t, d3.ID, parsed[2].DocID)
	})

	t.Run("invalid grades are filtered as missing", func(t *testing.T) {
		raw := `[
		  {"doc_id":"` + d1.ID.String() + `","grade":7},
		  {"doc_id":"` + d2.ID.String() + `","grade":-1},
		  {"doc_id":"` + d3.ID.String() + `","grade":2}
		]`
		parsed, missing, err := ParseBatchGradeJSON(raw, expected)
		require.NoError(t, err)
		assert.Len(t, parsed, 1)
		assert.Len(t, missing, 2)
	})

	t.Run("no array at all", func(t *testing.T) {
		_, missing, err := ParseBatchGradeJSON("Sorry, I cannot grade these.", expected)
		require.Error(t, err)
		assert.Len(t, missing, 3)
	})

	t.Run("malformed JSON inside array", func(t *testing.T) {
		_, missing, err := ParseBatchGradeJSON(`[{"oops":`, expected)
		require.Error(t, err)
		assert.Len(t, missing, 3)
	})
}
