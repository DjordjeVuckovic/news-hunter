package reader

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestYAMLMapper_Map(t *testing.T) {
	mapper := createMapper(t)

	articleID := uuid.New()
	published := time.Now().UTC().Truncate(time.Second)
	urlStr := "https://example.com"

	record := map[string]string{
		"id":        articleID.String(),
		"title":     "Test Article",
		"published": published.Format("2006-01-02T15:04:05Z"),
		"url":       urlStr,
	}

	article, err := mapper.Map(record, nil)
	require.NoError(t, err)

	assert.Equal(t, "Test Article", article.Title)
	assert.Equal(t, published, article.Metadata.PublishedAt)

	expectedURL, _ := url.Parse(urlStr)
	assert.Equal(t, *expectedURL, article.URL)
}

func createMapper(t *testing.T) *ArticleMapper {
	yamlContent := `
kind: DataMapping
version: v1
metadata:
  name: "Test"
dataset: test
fieldMappings:
  - source: "title"
    sourceType: "string"
    target: "Title"
    targetType: "string"
    required: true
  - source: "published"
    sourceType: "datetime"
    target: "Metadata.PublishedAt"
    targetType: "datetime"
    required: true
  - source: "url"
    sourceType: "url"
    target: "URL"
    targetType: "url"
    required: true
dateFormat: "2006-01-02T15:04:05Z"
`
	reader := strings.NewReader(yamlContent)
	cfg, err := NewYAMLConfigLoader(reader).Load(false)

	if err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	return NewArticleMapper(cfg)
}

func TestCSVReader_ReadParallel_MalformedRow(t *testing.T) {
	csvData := `id,title,author
1,Go Concurrency,John Doe
2,OnlyTitle
3,Interfaces in Go,Alice`

	ctx := t.Context()
	reader := NewCSVReader(strings.NewReader(csvData))

	resultsChan, err := reader.ReadParallel(ctx, 2)
	require.NoError(t, err)

	var validResults []map[string]string
	var errorCount int

	for res := range resultsChan {
		if res.Err != nil {
			errorCount++
			continue
		}
		validResults = append(validResults, res.Record)
	}

	// We expect one malformed line ("2,OnlyTitle") with missing columns
	assert.Equal(t, 2, len(validResults))     // 2 valid rows
	assert.Equal(t, 1, errorCount)            // 1 row caused an error
	assert.Contains(t, validResults[0], "id") // sanity check
}
