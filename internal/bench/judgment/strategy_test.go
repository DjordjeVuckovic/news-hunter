package judgment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStrategy_KnownNames(t *testing.T) {
	s, err := NewStrategy(StrategyLexical, StrategyOptions{})
	require.NoError(t, err)
	assert.Equal(t, "lexical", s.Name())

	s, err = NewStrategy(StrategyManual, StrategyOptions{})
	require.NoError(t, err)
	assert.Equal(t, "manual", s.Name())

	s, err = NewStrategy(StrategyBM25, StrategyOptions{})
	require.NoError(t, err)
	assert.Equal(t, "bm25", s.Name())
}

func TestNewStrategy_VectorHybridRequireEmbeddingEndpoint(t *testing.T) {
	// vector/hybrid are implemented but need an embedding backend; without one
	// they must fail loudly rather than silently producing no grades.
	for _, kind := range []StrategyKind{StrategyVector, StrategyHybrid} {
		_, err := NewStrategy(kind, StrategyOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding endpoint",
			"strategy %q should require an embedding endpoint", kind)
	}
}

func TestNewStrategy_VectorHybridWithEndpoint(t *testing.T) {
	for _, kind := range []StrategyKind{StrategyVector, StrategyHybrid} {
		s, err := NewStrategy(kind, StrategyOptions{EmbeddingBaseURL: "http://localhost:11434"})
		require.NoError(t, err)
		assert.Equal(t, string(kind), s.Name())
	}
}

func TestNewStrategy_UnknownName(t *testing.T) {
	_, err := NewStrategy("bogus", StrategyOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown strategy")
}
