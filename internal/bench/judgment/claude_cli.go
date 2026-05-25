package judgment

import (
	"context"
	"fmt"
	"os/exec"
)

const (
	defaultCLIBinary      = "claude"
	cliPreferredBatchSize = 10
)

// ClaudeCLIStrategy invokes the `claude -p <prompt>` CLI per (query, doc).
// One subprocess per call, focused single-doc context — keeps the model from
// drowning in a 13k-line enriched-pool YAML.
type ClaudeCLIStrategy struct {
	binary string
}

func NewClaudeCLIStrategy(opts StrategyOptions) *ClaudeCLIStrategy {
	bin := opts.CLIBinary
	if bin == "" {
		bin = defaultCLIBinary
	}
	return &ClaudeCLIStrategy{binary: bin}
}

func (s *ClaudeCLIStrategy) Name() string { return string(StrategyClaudeCLI) }

// ModelID returns the CLI binary name. The claude CLI's model is determined
// by its own runtime configuration (--model flag / user defaults) — we cannot
// reliably introspect it, so we record the binary identifier instead.
func (s *ClaudeCLIStrategy) ModelID() string { return s.binary }

func (s *ClaudeCLIStrategy) Grade(ctx context.Context, q GradingQuery, doc GradingDoc) (int, error) {
	out, err := s.runCLI(ctx, BuildGradingPrompt(q, doc))
	if err != nil {
		return 0, err
	}
	grade, err := ParseGradeJSON(out, doc.ID.String())
	if err != nil {
		return 0, fmt.Errorf("doc %s: %w", doc.ID, err)
	}
	return grade, nil
}

func (s *ClaudeCLIStrategy) PreferredBatchSize() int { return cliPreferredBatchSize }

// GradeBatch implements BatchStrategy. The CLI uses the same prompt structure
// as the API path, but the system prompt is prepended to the user payload —
// `claude -p` only accepts a single positional message.
func (s *ClaudeCLIStrategy) GradeBatch(ctx context.Context, q GradingQuery, docs []GradingDoc) ([]GradedDoc, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	prompt := BatchSystemPrompt + "\n\n" + BuildBatchGradingPrompt(q, docs)
	out, err := s.runCLI(ctx, prompt)
	if err != nil {
		return nil, err
	}
	parsed, missing, err := ParseBatchGradeJSON(out, docs)
	if err != nil {
		return nil, fmt.Errorf("batch parse: %w", err)
	}
	if len(missing) > 0 {
		return parsed, &PartialBatchError{Missing: missing, Got: len(parsed), Want: len(docs)}
	}
	return parsed, nil
}

func (s *ClaudeCLIStrategy) runCLI(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, s.binary, "-p", prompt)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s -p exit %d: %s", s.binary, exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("%s -p: %w", s.binary, err)
	}
	return string(out), nil
}
