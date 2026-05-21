package suite

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type TestSuite struct {
	SchemaVersion int              `yaml:"schema_version"`
	ID            string           `yaml:"id"`
	Name          string           `yaml:"name,omitempty"`
	Description   string           `yaml:"description,omitempty"`
	Version       string           `yaml:"version"`
	Corpus        *Corpus          `yaml:"corpus,omitempty"`
	Templates     []*QueryTemplate `yaml:"templates,omitempty"`
	Queries       []Query          `yaml:"queries"`
}

// Corpus records the dataset the suite targets. Lets a report attest "this
// was scored against the news_hunter_articles index, snapshot 2026-05-10".
type Corpus struct {
	Name       string `yaml:"name"`
	Source     string `yaml:"source,omitempty"`
	SnapshotAt string `yaml:"snapshot_at,omitempty"`
}

type Query struct {
	ID          string                 `yaml:"id"`
	Description string                 `yaml:"description"`
	Engines     map[string]EngineQuery `yaml:"engines"`
	Judgments   []RelevanceJudgment    `yaml:"judgments"`
}

type EngineQuery struct {
	Query    string         `yaml:"query,omitempty"`
	File     string         `yaml:"file,omitempty"`
	Template string         `yaml:"template,omitempty"`
	Params   TemplateParams `yaml:"params,omitempty"`
}

func (eq *EngineQuery) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		eq.Query = value.Value
		return nil
	}
	type plain EngineQuery
	return value.Decode((*plain)(eq))
}

func (eq *EngineQuery) Resolve(registry *TemplateRegistry, suiteDir string) (*ResolvedQuery, error) {
	if eq.Template != "" {
		if registry == nil {
			return nil, fmt.Errorf("template %q referenced but no registry available", eq.Template)
		}
		return registry.RenderQuery(eq.Template, eq.Params, suiteDir)
	}
	if eq.File != "" {
		path := eq.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(suiteDir, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read query file %q: %w", eq.File, err)
		}
		return &ResolvedQuery{Query: string(data)}, nil
	}
	return &ResolvedQuery{Query: eq.Query}, nil
}

type ResolvedQuery struct {
	Query string
}

type RelevanceJudgment struct {
	DocID     uuid.UUID `yaml:"doc_id"`
	Relevance int       `yaml:"relevance"`
}

func (q *Query) JudgmentMap() map[uuid.UUID]int {
	m := make(map[uuid.UUID]int, len(q.Judgments))
	for _, j := range q.Judgments {
		m[j.DocID] = j.Relevance
	}
	return m
}

// InjectJudgments sets the per-query Judgments slice from a flat map produced
// by the CLI layer after loading an annotations file. Replaces the loader-side
// auto-injection of v0; keeps the suite YAML focused on queries only.
func (ls *LoadedSuite) InjectJudgments(byQuery map[string][]RelevanceJudgment) {
	for i := range ls.Suite.Queries {
		if js, ok := byQuery[ls.Suite.Queries[i].ID]; ok {
			ls.Suite.Queries[i].Judgments = js
		}
	}
}

func (q *Query) ResolveEngineQuery(engine string, registry *TemplateRegistry, suiteDir string) (*ResolvedQuery, error) {
	eq, ok := q.Engines[engine]
	if !ok {
		return nil, nil
	}
	return eq.Resolve(registry, suiteDir)
}
