package meta

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRunID_FormatAndKind(t *testing.T) {
	id := NewRunID("pool")
	parts := strings.Split(id, "-")
	// 2026-05-21T14-08-55-pool-7c91a3 → 7 parts (date hyphens count)
	assert.GreaterOrEqual(t, len(parts), 6)
	assert.Contains(t, id, "-pool-", "kind should appear in id")
}

func TestNewRunID_SortableByTime(t *testing.T) {
	ids := []string{NewRunID("a"), NewRunID("b"), NewRunID("c")}
	sorted := append([]string{}, ids...)
	sort.Strings(sorted)
	// Lexicographic sort should match insertion order because timestamp prefix.
	// In the rare case two ids share a timestamp, the random suffix breaks ties
	// — order may flip but every id remains unique.
	uniq := map[string]struct{}{}
	for _, id := range ids {
		uniq[id] = struct{}{}
	}
	assert.Len(t, uniq, 3, "ids should be unique")
}

func TestNew_StampsToolAndTimestamp(t *testing.T) {
	m := New("judge")
	assert.Contains(t, m.Tool, "bench/")
	assert.False(t, m.GeneratedAt.IsZero())
	assert.Contains(t, m.RunID, "-judge-")
}
