package reader

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ArticleReflectTest struct {
	ID          uuid.UUID
	Title       string
	URL         url.URL
	Published   time.Time
	ViewCount   int64
	Rating      float64
	IsPublished bool
	CreatedDate time.Time
	Meta        struct {
		PublishedAt time.Time
		Priority    int
		Score       float32
		Active      bool
	}
}

func TestSetFlatField(t *testing.T) {
	var article ArticleReflectTest
	val := reflect.ValueOf(&article).Elem()

	// Test string
	err := SetFlatField(val, "Title", "Concurrency in Go", "string", "")
	require.NoError(t, err)
	assert.Equal(t, "Concurrency in Go", article.Title)

	// Test UUID
	err = SetFlatField(val, "ID", "123e4567-e89b-12d3-a456-426614174000", "uuid", "")
	require.NoError(t, err)

	// Test URL
	err = SetFlatField(val, "URL", "https://example.com", "url", "")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", article.URL.String())

	// Test datetime
	err = SetFlatField(val, "Published", "2024-10-01T12:00:00Z", "datetime", time.RFC3339)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC), article.Published)

	// Test int
	err = SetFlatField(val, "ViewCount", "42", "int", "")
	require.NoError(t, err)
	assert.Equal(t, int64(42), article.ViewCount)

	// Test float
	err = SetFlatField(val, "Rating", "4.5", "float", "")
	require.NoError(t, err)
	assert.Equal(t, 4.5, article.Rating)

	// Test bool
	err = SetFlatField(val, "IsPublished", "true", "bool", "")
	require.NoError(t, err)
	assert.Equal(t, true, article.IsPublished)

	// Test date
	err = SetFlatField(val, "CreatedDate", "2024-10-01", "date", "")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC), article.CreatedDate)
}

func TestSetNestedField(t *testing.T) {
	var article ArticleReflectTest
	val := reflect.ValueOf(&article).Elem()

	// Test datetime
	err := SetNestedField(val, []string{"Meta", "PublishedAt"}, "2024-10-01T12:00:00Z", "datetime", time.RFC3339)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC), article.Meta.PublishedAt)

	// Test int
	err = SetNestedField(val, []string{"Meta", "Priority"}, "5", "int", "")
	require.NoError(t, err)
	assert.Equal(t, 5, article.Meta.Priority)

	// Test float
	err = SetNestedField(val, []string{"Meta", "Score"}, "3.14", "float", "")
	require.NoError(t, err)
	assert.Equal(t, float32(3.14), article.Meta.Score)

	// Test bool
	err = SetNestedField(val, []string{"Meta", "Active"}, "false", "bool", "")
	require.NoError(t, err)
	assert.Equal(t, false, article.Meta.Active)
}

func TestFlexibleDateTimeParsing(t *testing.T) {
	var article ArticleReflectTest
	val := reflect.ValueOf(&article).Elem()

	// Test with microseconds format
	err := SetFlatField(val, "Published", "2023-10-30 10:12:35.000000", "datetime", "2006-01-02 15:04:05.000000")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2023, 10, 30, 10, 12, 35, 0, time.UTC), article.Published)

	// Test without microseconds format (should fallback automatically)
	err = SetFlatField(val, "CreatedDate", "2023-11-27 16:52:44", "datetime", "2006-01-02 15:04:05.000000")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2023, 11, 27, 16, 52, 44, 0, time.UTC), article.CreatedDate)

	// Test nested field with different format
	err = SetNestedField(val, []string{"Meta", "PublishedAt"}, "2023-11-27 22:20:41", "datetime", "2006-01-02 15:04:05.000000")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2023, 11, 27, 22, 20, 41, 0, time.UTC), article.Meta.PublishedAt)
}
