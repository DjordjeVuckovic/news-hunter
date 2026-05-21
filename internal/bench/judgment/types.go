package judgment

import (
	"github.com/google/uuid"
)

const (
	GradeUnjudged   = -1
	GradeNotRelev   = 0
	GradeMarginally = 1
	GradeRelevant   = 2
	GradeHighly     = 3
)

type GradedDoc struct {
	DocID uuid.UUID `yaml:"doc_id"`
	Grade int       `yaml:"grade"`
}

type JudgmentFile struct {
	Strategy string          `yaml:"strategy"`
	Queries  []JudgmentEntry `yaml:"queries"`
}

type JudgmentEntry struct {
	QueryID string      `yaml:"query_id"`
	Docs    []GradedDoc `yaml:"docs"`
}
