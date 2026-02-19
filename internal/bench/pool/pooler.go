package pool

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/bench/engine"
	"github.com/google/uuid"
)

type PoolFile struct {
	SuiteName string      `yaml:"suite_name"`
	Queries   []PoolEntry `yaml:"queries"`
}

type PoolEntry struct {
	QueryID   string      `yaml:"query_id"`
	QueryDesc string      `yaml:"query_desc"`
	Docs      []PooledDoc `yaml:"docs"`
}

type PooledDoc struct {
	DocID   uuid.UUID `yaml:"doc_id"`
	Sources []string  `yaml:"sources"`
}

func PoolResults(results map[string]*engine.Execution, depth int) []PooledDoc {
	seen := make(map[uuid.UUID]*PooledDoc)
	var order []uuid.UUID

	for engineName, exec := range results {
		if exec == nil {
			continue
		}
		limit := depth
		if limit > len(exec.RankedDocIDs) {
			limit = len(exec.RankedDocIDs)
		}
		for _, docID := range exec.RankedDocIDs[:limit] {
			if pd, ok := seen[docID]; ok {
				pd.Sources = append(pd.Sources, engineName)
			} else {
				seen[docID] = &PooledDoc{
					DocID:   docID,
					Sources: []string{engineName},
				}
				order = append(order, docID)
			}
		}
	}

	docs := make([]PooledDoc, 0, len(order))
	for _, id := range order {
		docs = append(docs, *seen[id])
	}
	return docs
}
