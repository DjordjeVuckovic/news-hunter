// Package embedfile reads precomputed document embeddings produced offline
// (see scripts/embed_qwen3.ipynb) from a Parquet file with columns
// `id` (string, article UUID) and `embedding` (list<float32>), plus file-level
// metadata carrying the model name and vector dimension.
package embedfile

import (
	"fmt"
	"os"
	"strconv"

	"github.com/parquet-go/parquet-go"
)

// Meta holds the file-level provenance written by the embedding notebook.
type Meta struct {
	Model      string
	Dim        int
	Pooling    string
	Normalized string
	RowCount   int
	CreatedAt  string
}

// Record is a single decoded embedding row.
type Record struct {
	ID        string
	Embedding []float32
}

type parquetRow struct {
	ID        string    `parquet:"id"`
	Embedding []float32 `parquet:"embedding"`
}

// Reader streams Records from a Parquet embeddings file.
type Reader struct {
	file *os.File
	pr   *parquet.GenericReader[parquetRow]
	meta Meta
}

// Open opens the Parquet file at path and reads its metadata block.
func Open(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open embeddings file: %w", err)
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("stat embeddings file: %w", err)
	}

	pf, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("parse parquet file: %w", err)
	}

	return &Reader{
		file: f,
		pr:   parquet.NewGenericReader[parquetRow](f),
		meta: metaFrom(pf),
	}, nil
}

// Meta returns the file-level metadata.
func (r *Reader) Meta() Meta { return r.meta }

// Read fills buf with up to len(buf) records, returning the count read.
// It returns io.EOF when the file is exhausted (possibly with n > 0).
func (r *Reader) Read(buf []Record) (int, error) {
	// Fresh slice per call: the decoded float slices are handed out to callers
	// (and retained in embedding.Vec), so they must not be reused across batches.
	rows := make([]parquetRow, len(buf))
	n, err := r.pr.Read(rows)
	for i := 0; i < n; i++ {
		buf[i] = Record{ID: rows[i].ID, Embedding: rows[i].Embedding}
	}
	return n, err
}

func (r *Reader) Close() error {
	r.pr.Close()
	return r.file.Close()
}

func metaFrom(pf *parquet.File) Meta {
	lookup := func(key string) string {
		v, _ := pf.Lookup(key)
		return v
	}
	dim, _ := strconv.Atoi(lookup("dim"))
	rowCount, _ := strconv.Atoi(lookup("row_count"))
	return Meta{
		Model:      lookup("model"),
		Dim:        dim,
		Pooling:    lookup("pooling"),
		Normalized: lookup("normalized"),
		RowCount:   rowCount,
		CreatedAt:  lookup("created_at"),
	}
}
