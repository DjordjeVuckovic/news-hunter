package pool

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func WritePoolFile(pf *PoolFile, path string) error {
	data, err := yaml.Marshal(pf)
	if err != nil {
		return fmt.Errorf("marshal pool file: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write pool file: %w", err)
	}
	return nil
}
