package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	dquery "github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"github.com/labstack/echo/v4"
)

type stubSemanticSearcher struct{}

func (stubSemanticSearcher) SearchSemantic(context.Context, *dquery.Semantic, *dquery.BaseOptions) (*storage.VectorSearchResult, error) {
	return &storage.VectorSearchResult{}, nil
}

func TestCapabilitiesHandler(t *testing.T) {
	tests := []struct {
		name         string
		withSemantic bool
		wantSemantic bool
	}{
		{name: "without semantic searcher", withSemantic: false, wantSemantic: false},
		{name: "with semantic searcher", withSemantic: true, wantSemantic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SearchRouter{e: echo.New()}
			if tt.withSemantic {
				r.semanticSearcher = stubSemanticSearcher{}
			}

			req := httptest.NewRequest(http.MethodGet, "/v1/capabilities", nil)
			rec := httptest.NewRecorder()
			c := r.e.NewContext(req, rec)

			if err := r.capabilitiesHandler(c); err != nil {
				t.Fatalf("capabilitiesHandler returned error: %v", err)
			}

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			var got map[string]bool
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to decode body: %v", err)
			}

			for _, key := range []string{"string_query", "match", "multi_match", "phrase", "boolean"} {
				if !got[key] {
					t.Errorf("%q = false, want true", key)
				}
			}
			if got["semantic"] != tt.wantSemantic {
				t.Errorf("semantic = %v, want %v", got["semantic"], tt.wantSemantic)
			}
		})
	}
}
