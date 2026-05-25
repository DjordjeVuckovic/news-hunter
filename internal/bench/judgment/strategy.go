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

// ModelIdentifier is an optional capability for strategies that know which
// specific model they used. cmd_judge checks this via type assertion to
// populate meta.JudgeModel accurately without relying on the --api-model flag.
// Deterministic strategies (lexical, manual) do not implement this interface.
type ModelIdentifier interface {
	ModelID() string
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

// Strategy taxonomy. Heuristic strategies derive grades from algorithmic
// signals; LLM strategies call out to a model; manual is a placeholder for
// human grading.
const (
	// Heuristic.
	StrategyLexical StrategyKind = "lexical" // token-overlap baseline
	StrategyBM25    StrategyKind = "bm25"    // reserved (not implemented)
	StrategyVector  StrategyKind = "vector"  // reserved (not implemented)
	StrategyHybrid  StrategyKind = "hybrid"  // reserved (not implemented)

	// LLM.
	StrategyClaudeCLI StrategyKind = "claude-cli"
	StrategyClaudeAPI StrategyKind = "claude-api"

	// Human.
	StrategyManual StrategyKind = "manual"
)

// KnownStrategies returns every strategy kind the runner can instantiate
// (including reserved ones). Used by the spec validator to reject
// spec.defaults.judgments typos at load time.
func KnownStrategies() []string {
	return []string{
		string(StrategyLexical),
		string(StrategyBM25),
		string(StrategyVector),
		string(StrategyHybrid),
		string(StrategyClaudeCLI),
		string(StrategyClaudeAPI),
		string(StrategyManual),
	}
}

type StrategyOptions struct {
	APIKey      string
	APIModel    string
	APIBaseURL  string
	CLIBinary   string
	Concurrency int
}

func NewStrategy(kind StrategyKind, opts StrategyOptions) (Strategy, error) {
	switch kind {
	case StrategyLexical:
		return NewLexicalStrategy(), nil
	case StrategyClaudeCLI:
		return NewClaudeCLIStrategy(opts), nil
	case StrategyClaudeAPI:
		return NewClaudeAPIStrategy(opts)
	case StrategyManual:
		return NewManualStrategy(), nil
	case StrategyBM25, StrategyVector, StrategyHybrid:
		return nil, fmt.Errorf("strategy %q is reserved but not yet implemented", kind)
	default:
		return nil, fmt.Errorf("unknown strategy %q (known: lexical, claude-cli, claude-api, manual)", kind)
	}
}
