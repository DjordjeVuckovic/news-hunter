package es

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// VectorStore is the Elasticsearch implementation of storage.VectorStore.
// It is a stub: the ES articles index has no dense_vector field yet, so vector
// retrieval is not wired. Postgres is the precedence engine for embeddings;
// this fills in once ES embeddings land. Mirrors the other "not yet
// implemented" ES stubs in storage/factory.
type VectorStore struct{}

func NewVectorStore() *VectorStore { return &VectorStore{} }

func (*VectorStore) QueryVector(context.Context, string) ([]float32, error) {
	return nil, fmt.Errorf("es vector store not yet implemented")
}

func (*VectorStore) DocVectors(context.Context, []uuid.UUID) (map[uuid.UUID][]float32, error) {
	return nil, fmt.Errorf("es vector store not yet implemented")
}
