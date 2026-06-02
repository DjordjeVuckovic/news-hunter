package embedfile

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/parquet-go/parquet-go"
)

func writeFixture(t *testing.T, rows []parquetRow, meta map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "embeddings.parquet")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	defer f.Close()

	kv := make([]parquet.WriterOption, 0, len(meta))
	for k, v := range meta {
		kv = append(kv, parquet.KeyValueMetadata(k, v))
	}

	w := parquet.NewGenericWriter[parquetRow](f, kv...)
	if _, err := w.Write(rows); err != nil {
		t.Fatalf("write rows: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return path
}

func TestReadRecordsAndMeta(t *testing.T) {
	rows := []parquetRow{
		{ID: "11111111-1111-1111-1111-111111111111", Embedding: []float32{0.1, 0.2, 0.3}},
		{ID: "22222222-2222-2222-2222-222222222222", Embedding: []float32{0.4, 0.5, 0.6}},
		{ID: "33333333-3333-3333-3333-333333333333", Embedding: []float32{0.7, 0.8, 0.9}},
	}
	path := writeFixture(t, rows, map[string]string{
		"model":     "qwen3-embedding:0.6b",
		"dim":       "3",
		"pooling":   "last_token",
		"row_count": "3",
	})

	r, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	meta := r.Meta()
	if meta.Model != "qwen3-embedding:0.6b" {
		t.Errorf("model = %q, want qwen3-embedding:0.6b", meta.Model)
	}
	if meta.Dim != 3 {
		t.Errorf("dim = %d, want 3", meta.Dim)
	}
	if meta.RowCount != 3 {
		t.Errorf("row_count = %d, want 3", meta.RowCount)
	}

	var got []Record
	buf := make([]Record, 2) // smaller than total to exercise batching
	for {
		n, err := r.Read(buf)
		got = append(got, buf[:n]...)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
	}

	if len(got) != len(rows) {
		t.Fatalf("read %d records, want %d", len(got), len(rows))
	}
	for i, rec := range got {
		if rec.ID != rows[i].ID {
			t.Errorf("record %d id = %q, want %q", i, rec.ID, rows[i].ID)
		}
		if len(rec.Embedding) != len(rows[i].Embedding) {
			t.Fatalf("record %d dim = %d, want %d", i, len(rec.Embedding), len(rows[i].Embedding))
		}
		for j, v := range rec.Embedding {
			if v != rows[i].Embedding[j] {
				t.Errorf("record %d[%d] = %v, want %v", i, j, v, rows[i].Embedding[j])
			}
		}
	}
}
