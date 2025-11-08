package reader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLMapper_String_LoadConfig(t *testing.T) {
	// Arrange: prepare YAML content and write it to a test file
	reader := strings.NewReader(`
kind: DataMapping
version: v1
metadata:
  name: "Kaggle Dataset"
dataset: kaggle
fieldMappings:
  - source: "id"
    target: "id"
`)
	loader := NewYAMLConfigLoader(reader)

	// Act
	cfg, err := loader.Load(false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "v1", cfg.Version)
	assert.Equal(t, "DataMapping", cfg.Kind)
	assert.Equal(t, "Kaggle Dataset", cfg.Metadata.Name)
	assert.Equal(t, "kaggle", cfg.Dataset)
	assert.Len(t, cfg.FieldMappings, 1)
	assert.Equal(t, "id", cfg.FieldMappings[0].Source)
}

func TestYAMLMapper_File_LoadConfig(t *testing.T) {
	// Arrange: prepare YAML content and write it to a test file
	path, err := createTestFile(t, true)
	require.NoError(t, err)

	fileReader := func() (*os.File, error) {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	reader, err := fileReader()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Logf("failed to remove test file: %v", err)
		}
	}(path)
	require.NoError(t, err)

	loader := NewYAMLConfigLoader(reader)

	// Act
	cfg, err := loader.Load(false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "v1", cfg.Version)
	assert.Equal(t, "DataMapping", cfg.Kind)
	assert.Equal(t, "Kaggle Dataset", cfg.Metadata.Name)
	assert.Equal(t, "kaggle", cfg.Dataset)
	assert.Len(t, cfg.FieldMappings, 1)
	assert.Equal(t, "id", cfg.FieldMappings[0].Source)
}

func TestYAMLMapper_LoadConfig_ShouldFail(t *testing.T) {
	// Arrange
	path, err := createTestFile(t, false)
	require.NoError(t, err)

	fileReader := func() (*os.File, error) {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	reader, err := fileReader()
	defer os.Remove(path)
	require.NoError(t, err)

	loader := NewYAMLConfigLoader(reader)

	// Act
	cfg, err := loader.Load(false)

	// Act
	if err != nil {
		t.Logf("Expected error: %v", err)
	}
	assert.Equal(t, 0, len(cfg.FieldMappings))
}

func createTestFile(t *testing.T, validSyntax bool) (string, error) {
	if !validSyntax {
		invalidYAMLContent := `
kind: DataMapping
version: v1
metadata:
  name: "Invalid Mapping"
dataset: kaggle
field_mappings:
 - source: "title"
   sourceType: "string"
   target: "Title"
   targetType: "string"
dateFormat: "2006-01-02T15:04:05Z"
`
		// Create a temp test file in relative path (e.g. ./testdata/)
		dir := "testdata"
		name := "invalid-mapping.yaml"
		err := os.MkdirAll(dir, 0755)

		if err != nil {
			return "", err
		}
		filePath := filepath.Join(dir, name)

		err = os.WriteFile(filePath, []byte(invalidYAMLContent), 0644)
		if err != nil {
			return "", err
		}
		require.NoError(t, err)
		return filePath, nil
	}
	validYAMLContent := `
kind: DataMapping
version: v1
metadata:
  name: "Kaggle Dataset"
dataset: kaggle
fieldMappings:
  - source: "id"
    target: "id"
`
	// Create a temp test file in relative path (e.g. ./testdata/)
	dir := "testdata"
	name := "valid-mapping.yaml"
	err := os.MkdirAll(dir, 0755)

	require.NoError(t, err)
	filePath := filepath.Join(dir, name)
	err = os.WriteFile(filePath, []byte(validYAMLContent), 0644)
	require.NoError(t, err)

	return filePath, nil
}
