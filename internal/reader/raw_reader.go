package reader

import "context"

type Reader interface {
	Read() ([]map[string]string, error)
}
type ParallelReaderResult struct {
	Record map[string]string
	Err    error
}
type RawParallelReader interface {
	ReadParallel(ctx context.Context, workerCount int) (<-chan ParallelReaderResult, error)
}
