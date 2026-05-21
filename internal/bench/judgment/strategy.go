package judgment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// GradingQuery is the search intent the judge uses to grade.
type GradingQuery struct {
	ID          string
	Description string
}

// GradingDoc is a single article candidate to grade.
type GradingDoc struct {
	ID          uuid.UUID
	Title       string
	Description string
	Content     string
}

// Strategy grades one (query, doc) pair at a time. Implementations must be
// safe for concurrent calls — the runner dispatches multiple goroutines.
type Strategy interface {
	Name() string
	Grade(ctx context.Context, q GradingQuery, doc GradingDoc) (int, error)
}

// BatchStrategy is an optional capability: strategies that can grade N docs
// in a single LLM call should implement it. The runner detects it via type
// assertion and prefers GradeBatch over Grade when present.
//
// The pattern follows Anthropic's "LLM as judge — batched" cookbook:
//   - one system prompt sets the rubric
//   - one user message containing the query + numbered candidates
//   - response is a single JSON array, one entry per candidate
//
// Implementations MUST tolerate partial responses: if the model returns N-k
// entries, return what was parsed and let the runner re-dispatch the missing
// docs through Grade() as a fallback.
type BatchStrategy interface {
	Strategy
	PreferredBatchSize() int
	GradeBatch(ctx context.Context, q GradingQuery, docs []GradingDoc) ([]GradedDoc, error)
}

type StrategyKind string

const (
	StrategyKeyword   StrategyKind = "keyword"
	StrategyClaudeCLI StrategyKind = "claude-cli"
	StrategyClaudeAPI StrategyKind = "claude-api"
	StrategyStub      StrategyKind = "stub"
)

type StrategyOptions struct {
	APIKey      string
	APIModel    string
	APIBaseURL  string
	CLIBinary   string
	Concurrency int
}

func NewStrategy(kind StrategyKind, opts StrategyOptions) (Strategy, error) {
	switch kind {
	case StrategyKeyword:
		return NewKeywordStrategy(), nil
	case StrategyClaudeCLI:
		return NewClaudeCLIStrategy(opts), nil
	case StrategyClaudeAPI:
		return NewClaudeAPIStrategy(opts)
	case StrategyStub:
		return NewStubStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown strategy %q", kind)
	}
}
