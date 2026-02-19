package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("valid spec", func(t *testing.T) {
		yaml := `
engines:
  pg-native:
    type: postgres
    connection: "postgresql://localhost/test"
  elasticsearch:
    type: elasticsearch
    connection: "http://localhost:9200"
    index: news

metrics:
  k_values: [3, 5, 10]
  max_k: 100
  relevance_threshold: 1

runs:
  warmup: 1
  iterations: 3

jobs:
  - name: "raw-comparison"
    suite: configs/bench/suite.yaml
    engines: [pg-native, elasticsearch]
`
		s, err := Parse([]byte(yaml))
		require.NoError(t, err)
		assert.Len(t, s.Jobs, 1)
		assert.Len(t, s.Engines, 2)
		assert.Equal(t, "raw-comparison", s.Jobs[0].Name)
		assert.Equal(t, 3, s.Runs.Iterations)
	})

	t.Run("no jobs", func(t *testing.T) {
		yaml := `
engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs: []
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no jobs")
	})

	t.Run("no engines", func(t *testing.T) {
		yaml := `
engines: {}
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg]
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no engines")
	})

	t.Run("job references unknown engine", func(t *testing.T) {
		yaml := `
engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg, unknown]
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown engine")
	})

	t.Run("defaults applied", func(t *testing.T) {
		yaml := `
engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
metrics: {}
runs: {}
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg]
`
		s, err := Parse([]byte(yaml))
		require.NoError(t, err)
		assert.Equal(t, 100, s.Metrics.MaxK)
		assert.Equal(t, []int{3, 5, 10}, s.Metrics.KValues)
		assert.Equal(t, 1, s.Metrics.RelevanceThreshold)
		assert.Equal(t, 1, s.Runs.Iterations)
	})

	t.Run("invalid engine type", func(t *testing.T) {
		yaml := `
engines:
  bad:
    type: mysql
    connection: "mysql://localhost"
jobs:
  - name: test
    suite: suite.yaml
    engines: [bad]
`
		_, err := Parse([]byte(yaml))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("api engine type is valid", func(t *testing.T) {
		yaml := `
engines:
  api:
    type: api
    connection: "http://localhost:8080"
jobs:
  - name: api-test
    suite: suite.yaml
    engines: [api]
`
		s, err := Parse([]byte(yaml))
		require.NoError(t, err)
		assert.Equal(t, "api", s.Engines["api"].Type)
		assert.Equal(t, "http://localhost:8080", s.Engines["api"].Connection)
	})
}
