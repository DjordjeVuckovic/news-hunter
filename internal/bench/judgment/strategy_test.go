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
}

func TestNewStrategy_ReservedNotImplemented(t *testing.T) {
	for _, kind := range []StrategyKind{StrategyBM25, StrategyVector, StrategyHybrid} {
		_, err := NewStrategy(kind, StrategyOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reserved but not yet implemented",
			"reserved strategy %q should fail loudly", kind)
	}
}

func TestNewStrategy_UnknownName(t *testing.T) {
	_, err := NewStrategy("bogus", StrategyOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown strategy")
}
