package pool

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadPoolFile(path string) (*PoolFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pool file: %w", err)
	}
	var pf PoolFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parse pool file: %w", err)
	}
	return &pf, nil
}
