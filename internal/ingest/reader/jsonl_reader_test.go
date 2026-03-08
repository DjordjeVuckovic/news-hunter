package reader

import (
	"context"
	"strings"
	"testing"
)

const sampleJSONL = `{"id":"abc123","title":"Test Article","author":"Jane Doe","content":"Some content"}
{"id":"def456","title":"Second Article","author":"","content":"More content"}
`

func TestJSONLReader_Read(t *testing.T) {
	jr := NewJSONLReader(strings.NewReader(sampleJSONL))
	records, err := jr.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0]["id"] != "abc123" {
		t.Errorf("expected id abc123, got %s", records[0]["id"])
	}
	if records[0]["title"] != "Test Article" {
		t.Errorf("expected title 'Test Article', got %s", records[0]["title"])
	}
	if records[1]["id"] != "def456" {
		t.Errorf("expected id def456, got %s", records[1]["id"])
	}
}

func TestJSONLReader_ReadParallel(t *testing.T) {
	jr := NewJSONLReader(strings.NewReader(sampleJSONL))
	ctx := context.Background()
	ch, err := jr.ReadParallel(ctx, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []ParallelReaderResult
	for r := range ch {
		results = append(results, r)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected record error: %v", r.Err)
		}
		if r.Record["id"] == "" {
			t.Error("expected non-empty id")
		}
	}
}

func TestJSONLReader_SkipsNullValues(t *testing.T) {
	input := `{"id":"x1","title":"Article","nullable":null}` + "\n"
	jr := NewJSONLReader(strings.NewReader(input))
	records, err := jr.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := records[0]["nullable"]; exists {
		t.Error("null field should be skipped")
	}
}
