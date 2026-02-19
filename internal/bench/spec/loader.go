package spec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadFromFile(path string) (*BenchSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (*BenchSpec, error) {
	var s BenchSpec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse spec YAML: %w", err)
	}
	if err := validate(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

var validEngineTypes = map[string]bool{
	"postgres":      true,
	"elasticsearch": true,
	"api":           true,
}

func validate(s *BenchSpec) error {
	if len(s.Jobs) == 0 {
		return fmt.Errorf("spec has no jobs")
	}
	if len(s.Engines) == 0 {
		return fmt.Errorf("spec has no engines")
	}
	for i, j := range s.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job at index %d has no name", i)
		}
		if j.Suite == "" {
			return fmt.Errorf("job %q has no suite", j.Name)
		}
		if len(j.Engines) == 0 {
			return fmt.Errorf("job %q has no engines", j.Name)
		}
		for _, engRef := range j.Engines {
			if _, ok := s.Engines[engRef]; !ok {
				return fmt.Errorf("job %q references unknown engine %q", j.Name, engRef)
			}
		}
	}
	for name, eng := range s.Engines {
		if eng.Type == "" {
			return fmt.Errorf("engine %q has no type", name)
		}
		if !validEngineTypes[eng.Type] {
			return fmt.Errorf("engine %q has invalid type %q", name, eng.Type)
		}
		if eng.Connection == "" {
			return fmt.Errorf("engine %q has no connection", name)
		}
	}
	if s.Metrics.MaxK <= 0 {
		s.Metrics.MaxK = 100
	}
	if len(s.Metrics.KValues) == 0 {
		s.Metrics.KValues = []int{3, 5, 10}
	}
	if s.Metrics.RelevanceThreshold <= 0 {
		s.Metrics.RelevanceThreshold = 1
	}
	if s.Runs.Iterations <= 0 {
		s.Runs.Iterations = 1
	}
	return nil
}
