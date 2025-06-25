package ingest

import "context"

// Pipeline defines the common interface for data ingestion pipelines
type Pipeline interface {
	// Run executes the pipeline with the given context
	Run(ctx context.Context) error

	// Stop gracefully stops the pipeline
	Stop()
}

// BulkOptions defines common bulk processing options
type BulkOptions struct {
	Enabled bool
	Size    int
}

// PipelineConfig defines common configuration for all pipelines
type PipelineConfig struct {
	Name string
	Bulk *BulkOptions
}
