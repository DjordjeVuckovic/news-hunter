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
