package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type EsExecutor struct {
	name    string
	baseURL string
	index   string
	client  *http.Client
}

func NewEsExecutor(name, baseURL, index string) *EsExecutor {
	return &EsExecutor{
		name:    name,
		baseURL: baseURL,
		index:   index,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *EsExecutor) Execute(ctx context.Context, rawQuery string, _ []any) (*Execution, error) {
	url := fmt.Sprintf("%s/%s/_search", e.baseURL, e.index)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(rawQuery))
	if err != nil {
		return nil, fmt.Errorf("es create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("es request: %w", err)
	}
	defer resp.Body.Close()
	latency := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("es read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("es status %d: %s", resp.StatusCode, string(body))
	}

	var esResp esSearchResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		return nil, fmt.Errorf("es parse response: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		id, err := uuid.Parse(hit.Source.ID)
		if err != nil {
			return nil, fmt.Errorf("es parse doc id %q: %w", hit.Source.ID, err)
		}
		ids = append(ids, id)
	}

	return &Execution{
		RankedDocIDs: ids,
		TotalMatches: esResp.Hits.Total.Value,
		Latency:      latency,
	}, nil
}

func (e *EsExecutor) Name() string { return e.name }
func (e *EsExecutor) Close() error { return nil }

// Validate posts the query to <index>/_validate/query?explain=true. ES parses
// the JSON, type-checks fields, and returns "valid: false" with an explanation
// for malformed queries — no documents scanned.
func (e *EsExecutor) Validate(ctx context.Context, rawQuery string) error {
	url := fmt.Sprintf("%s/%s/_validate/query?explain=true", e.baseURL, e.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(rawQuery))
	if err != nil {
		return fmt.Errorf("es validate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("es validate http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("es validate read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("es validate status %d: %s", resp.StatusCode, string(body))
	}

	var v esValidateResponse
	if err := json.Unmarshal(body, &v); err != nil {
		return fmt.Errorf("es validate parse: %w", err)
	}
	if !v.Valid {
		for _, exp := range v.Explanations {
			if exp.Error != "" {
				return fmt.Errorf("es invalid: %s", exp.Error)
			}
		}
		// _validate/query rejects top-level fields like "size","from","aggs".
		// Strip them and retry with just the query block before failing.
		stripped, ok := stripToQueryBody(rawQuery)
		if !ok {
			return fmt.Errorf("es invalid: %s", string(body))
		}
		return e.validateBody(ctx, stripped, body)
	}
	return nil
}

// validateBody is a second-chance validation against a stripped query body
// (just the `query` field). origBody is the original response, used in the
// error message if the retry also fails.
func (e *EsExecutor) validateBody(ctx context.Context, body []byte, origBody []byte) error {
	url := fmt.Sprintf("%s/%s/_validate/query?explain=true", e.baseURL, e.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("es validate retry: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("es validate retry http: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var v esValidateResponse
	if err := json.Unmarshal(respBody, &v); err != nil {
		return fmt.Errorf("es invalid: %s", string(origBody))
	}
	if !v.Valid {
		for _, exp := range v.Explanations {
			if exp.Error != "" {
				return fmt.Errorf("es invalid: %s", exp.Error)
			}
		}
		return fmt.Errorf("es invalid: %s", string(respBody))
	}
	return nil
}

// stripToQueryBody extracts just the "query" field from an ES search body —
// _validate/query rejects top-level "size", "from", "aggs", etc.
func stripToQueryBody(raw string) ([]byte, bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, false
	}
	q, ok := m["query"]
	if !ok {
		return nil, false
	}
	out, err := json.Marshal(map[string]json.RawMessage{"query": q})
	if err != nil {
		return nil, false
	}
	return out, true
}

type esValidateResponse struct {
	Valid        bool                    `json:"valid"`
	Explanations []esValidateExplanation `json:"explanations,omitempty"`
}

type esValidateExplanation struct {
	Index string `json:"index"`
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type esSearchResponse struct {
	Hits esHits `json:"hits"`
}

type esHits struct {
	Total esTotal `json:"total"`
	Hits  []esHit `json:"hits"`
}

type esTotal struct {
	Value int64 `json:"value"`
}

type esHit struct {
	Source esSource `json:"_source"`
}

type esSource struct {
	ID string `json:"id"`
}
