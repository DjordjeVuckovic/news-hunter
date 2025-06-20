package reader

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type Article struct {
	ID        uuid.UUID
	Title     string
	URL       url.URL
	Published time.Time
	Meta      struct {
		PublishedAt time.Time
	}
}

func TestSetFlatField(t *testing.T) {
	var article Article
	val := reflect.ValueOf(&article).Elem()

	err := SetFlatField(val, "Title", "Concurrency in Go", "string", "")
	require.NoError(t, err)

	err = SetFlatField(val, "ID", "123e4567-e89b-12d3-a456-426614174000", "uuid", "")
	require.NoError(t, err)

	err = SetFlatField(val, "URL", "https://example.com", "url", "")
	require.NoError(t, err)

	err = SetFlatField(val, "Published", "2024-10-01T12:00:00Z", "datetime", time.RFC3339)
	require.NoError(t, err)

	assert.Equal(t, "Concurrency in Go", article.Title)
	assert.Equal(t, "https://example.com", article.URL.String())
}

func TestSetNestedField(t *testing.T) {
	var article Article
	val := reflect.ValueOf(&article).Elem()

	err := SetNestedField(val, []string{"Meta", "PublishedAt"}, "2024-10-01T12:00:00Z", "datetime", time.RFC3339)
	require.NoError(t, err)

	assert.Equal(t, time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC), article.Meta.PublishedAt)
}
