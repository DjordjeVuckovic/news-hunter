package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSpecHeader = `schema_version: 1
id: test_spec
`

func TestParse(t *testing.T) {
	t.Run("valid spec", func(t *testing.T) {
		yaml := validSpecHeader + `engines:
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
    suite: tracks/test/suite.yaml
    engines: [pg-native, elasticsearch]
`
		s, err := Parse([]byte(yaml))
		require.NoError(t, err)
		assert.Equal(t, "test_spec", s.ID)
		assert.Len(t, s.Jobs, 1)
		assert.Len(t, s.Engines, 2)
		assert.Equal(t, "raw-comparison", s.Jobs[0].Name)
		assert.Equal(t, 3, s.Runs.Iterations)
	})

	t.Run("no jobs", func(t *testing.T) {
		yaml := validSpecHeader + `engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs: []
`
		_, err := Parse([]byte(yaml))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no jobs")
	})

	t.Run("no engines", func(t *testing.T) {
		yaml := validSpecHeader + `engines: {}
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg]
`
		_, err := Parse([]byte(yaml))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no engines")
	})

	t.Run("job references unknown engine", func(t *testing.T) {
		yaml := validSpecHeader + `engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg, unknown]
`
		_, err := Parse([]byte(yaml))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown engine")
	})

	t.Run("defaults applied", func(t *testing.T) {
		yaml := validSpecHeader + `engines:
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
		yaml := validSpecHeader + `engines:
  bad:
    type: mysql
    connection: "mysql://localhost"
jobs:
  - name: test
    suite: suite.yaml
    engines: [bad]
`
		_, err := Parse([]byte(yaml))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("api engine type is valid", func(t *testing.T) {
		yaml := validSpecHeader + `engines:
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

func TestParse_RejectsMissingSchemaVersion(t *testing.T) {
	yaml := `id: test
engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg]
`
	_, err := Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing schema_version")
}

func TestParse_RejectsMissingID(t *testing.T) {
	yaml := `schema_version: 1
engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg]
`
	_, err := Parse([]byte(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: id")
}

func TestParse_DefaultsBlock(t *testing.T) {
	yaml := validSpecHeader + `defaults:
  pool_depth: 50
  judgments: lexical
engines:
  pg:
    type: postgres
    connection: "postgresql://localhost/test"
jobs:
  - name: test
    suite: suite.yaml
    engines: [pg]
`
	s, err := Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, 50, s.Defaults.PoolDepth)
	assert.Equal(t, "lexical", s.Defaults.Judgments)
}
