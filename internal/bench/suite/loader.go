package suite

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type LoadedSuite struct {
	Suite    *TestSuite
	Registry *TemplateRegistry
	Dir      string
}

func LoadFromFile(path string) (*LoadedSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read suite file: %w", err)
	}
	loaded, err := Parse(data)
	if err != nil {
		return nil, err
	}
	loaded.Dir = filepath.Dir(path)

	if loaded.Suite.JudgmentsFile != "" {
		jfPath := loaded.Suite.JudgmentsFile
		if !filepath.IsAbs(jfPath) {
			jfPath = filepath.Join(loaded.Dir, jfPath)
		}
		if err := loaded.injectJudgmentsFromFile(jfPath); err != nil {
			return nil, fmt.Errorf("load judgments_file %q: %w", loaded.Suite.JudgmentsFile, err)
		}
	}
	return loaded, nil
}

type judgmentsYAML struct {
	Queries []struct {
		QueryID string `yaml:"query_id"`
		Docs    []struct {
			DocID uuid.UUID `yaml:"doc_id"`
			Grade int       `yaml:"grade"`
		} `yaml:"docs"`
	} `yaml:"queries"`
}

func (ls *LoadedSuite) injectJudgmentsFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		// A missing judgments file is the normal state during validate/pool/
		// judge. Stay silent here; 'bench run' surfaces the "no judgments"
		// condition in the report table when it matters.
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read judgments file: %w", err)
	}
	var jy judgmentsYAML
	if err := yaml.Unmarshal(data, &jy); err != nil {
		return fmt.Errorf("parse judgments file: %w", err)
	}

	byQuery := make(map[string][]RelevanceJudgment, len(jy.Queries))
	for _, qe := range jy.Queries {
		js := make([]RelevanceJudgment, 0, len(qe.Docs))
		for _, d := range qe.Docs {
			if d.Grade < 0 {
				continue
			}
			js = append(js, RelevanceJudgment{DocID: d.DocID, Relevance: d.Grade})
		}
		byQuery[qe.QueryID] = js
	}

	for i := range ls.Suite.Queries {
		if js, ok := byQuery[ls.Suite.Queries[i].ID]; ok {
			ls.Suite.Queries[i].Judgments = js
		}
	}
	return nil
}

func Parse(data []byte) (*LoadedSuite, error) {
	var s TestSuite
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse suite YAML: %w", err)
	}
	if len(s.Queries) == 0 {
		return nil, fmt.Errorf("suite has no queries")
	}

	registry := NewTemplateRegistry()
	for _, t := range s.Templates {
		if err := registry.Register(t); err != nil {
			return nil, fmt.Errorf("register template: %w", err)
		}
	}

	for i, q := range s.Queries {
		if q.ID == "" {
			return nil, fmt.Errorf("query at index %d has no id", i)
		}
		if len(q.Engines) == 0 {
			return nil, fmt.Errorf("query %q has no engines", q.ID)
		}
		for engName, eq := range q.Engines {
			if eq.Template != "" {
				if _, ok := registry.Get(eq.Template); !ok {
					return nil, fmt.Errorf("query %q engine %q references unknown template %q", q.ID, engName, eq.Template)
				}
			}
		}
	}

	return &LoadedSuite{Suite: &s, Registry: registry}, nil
}
