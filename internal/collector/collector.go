package collector

import "context"

type CollectionResult[T any] struct {
	Result T
	Err    error
}

type Collector[T any] interface {
	Collect(ctx context.Context) (<-chan CollectionResult[T], error)
}
