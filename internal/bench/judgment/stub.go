package judgment

import "context"

// StubStrategy emits GradeUnjudged (-1) for every doc. Useful when a human
// wants to grade from scratch — the runner writes a JudgmentFile with
// placeholders the user can fill in by hand.
type StubStrategy struct{}

func NewStubStrategy() *StubStrategy { return &StubStrategy{} }

func (StubStrategy) Name() string { return string(StrategyStub) }

func (StubStrategy) Grade(_ context.Context, _ GradingQuery, _ GradingDoc) (int, error) {
	return GradeUnjudged, nil
}
