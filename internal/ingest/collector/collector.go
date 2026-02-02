package collector

import "context"

type Result[T any] struct {
	Result T
	Err    error
}

type Collector[T any] interface {
	Collect(ctx context.Context) (<-chan Result[T], error)
}
