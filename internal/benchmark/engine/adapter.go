package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/benchmark/suite"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"github.com/google/uuid"
)

type QueryExecution struct {
	RankedDocIDs []uuid.UUID
	TotalMatches int64
	Latency      time.Duration
	Error        error
}

func Execute(ctx context.Context, eng SearchEngine, bq *suite.BenchmarkQuery, maxK int) QueryExecution {
	domainQuery, err := suite.ToDomainQuery(bq)
	if err != nil {
		return QueryExecution{Error: fmt.Errorf("convert query %q: %w", bq.ID, err)}
	}

	opts := &query.BaseOptions{Size: maxK}

	start := time.Now()
	result, err := dispatch(ctx, eng.Searcher, domainQuery, opts)
	latency := time.Since(start)

	if err != nil {
		return QueryExecution{Error: fmt.Errorf("execute query %q on %s: %w", bq.ID, eng.Name, err), Latency: latency}
	}

	return QueryExecution{
		RankedDocIDs: extractDocIDs(result),
		TotalMatches: result.TotalMatches,
		Latency:      latency,
	}
}

func dispatch(ctx context.Context, searcher storage.FtsSearcher, q *query.Base, opts *query.BaseOptions) (*storage.SearchResult, error) {
	switch q.Kind {
	case query.QueryStringType:
		return searcher.SearchQuery(ctx, q.QueryString, opts)
	case query.MatchType:
		return searcher.SearchField(ctx, q.Match, opts)
	case query.MultiMatchType:
		return searcher.SearchFields(ctx, q.MultiMatch, opts)
	case query.PhraseType:
		return searcher.SearchPhrase(ctx, q.Phrase, opts)
	case query.BooleanType:
		return searcher.SearchBoolean(ctx, q.Boolean, opts)
	default:
		return nil, fmt.Errorf("unsupported query kind: %s", q.Kind)
	}
}

func extractDocIDs(result *storage.SearchResult) []uuid.UUID {
	ids := make([]uuid.UUID, len(result.Hits))
	for i, hit := range result.Hits {
		ids[i] = hit.ID
	}
	return ids
}
