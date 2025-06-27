package processor

import (
	"context"
	"log/slog"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
)

const defaultBatchSize = 1000

// Pipeline defines the interface for data processing pipelines
type Pipeline interface {
	// Run executes the pipeline with the given context
	Run(ctx context.Context) error

	// Stop gracefully stops the pipeline
	Stop()
}

// BulkOptions defines bulk processing configuration
type BulkOptions struct {
	Enabled bool
	Size    int
}

// PipelineConfig defines configuration for pipelines
type PipelineConfig struct {
	Name string
	Bulk *BulkOptions
}

// ArticlePipeline handles article processing from collection to storage
type ArticlePipeline struct {
	collector collector.Collector[domain.Article]
	storer    storage.Storer
	config    *PipelineConfig
}

type PipelineOption func(pipeline *ArticlePipeline)

// WithBulk configures bulk processing with specified batch size
func WithBulk(size int) PipelineOption {
	return func(pipeline *ArticlePipeline) {
		if pipeline.config.Bulk == nil {
			pipeline.config.Bulk = &BulkOptions{}
		}
		pipeline.config.Bulk.Enabled = true
		pipeline.config.Bulk.Size = size
	}
}

// WithConfig sets custom pipeline configuration
func WithConfig(config *PipelineConfig) PipelineOption {
	return func(pipeline *ArticlePipeline) {
		pipeline.config = config
	}
}

// NewPipeline creates a new generic article processing pipeline
func NewPipeline(c collector.Collector[domain.Article], storer storage.Storer, opts ...PipelineOption) *ArticlePipeline {
	p := &ArticlePipeline{
		collector: c,
		storer:    storer,
		config: &PipelineConfig{
			Name: "article-pipeline",
			Bulk: &BulkOptions{
				Enabled: false,
				Size:    defaultBatchSize,
			},
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Run executes the pipeline
func (p *ArticlePipeline) Run(ctx context.Context) error {
	start := time.Now()
	slog.Info("🛫 Starting pipeline run",
		"pipeline", p.config.Name,
		"bulk_enabled", p.config.Bulk.Enabled,
		"batch_size", p.config.Bulk.Size,
		"time", start,
	)

	results, err := p.collector.Collect(ctx)
	if err != nil {
		slog.Error("Error collecting articles", "error", err, "pipeline", p.config.Name)
		return err
	}

	var runErr error
	if p.config.Bulk.Enabled {
		runErr = p.processBatch(ctx, results)
	} else {
		runErr = p.processBasic(ctx, results)
	}

	duration := time.Since(start)
	slog.Info("Pipeline run completed",
		"pipeline", p.config.Name,
		"duration", duration,
		"error", runErr,
	)

	return runErr
}

// processBasic handles individual article processing
func (p *ArticlePipeline) processBasic(ctx context.Context, results <-chan collector.Result[domain.Article]) error {
	processedCount := 0
	errorCount := 0

	for {
		select {
		case <-ctx.Done():
			slog.Info("Pipeline context cancelled, stopping collection",
				"pipeline", p.config.Name,
				"processed", processedCount,
				"errors", errorCount,
			)
			return ctx.Err()
		case res, ok := <-results:
			if !ok {
				slog.Info("Collection channel closed, stopping collection",
					"pipeline", p.config.Name,
					"processed", processedCount,
					"errors", errorCount,
				)
				return nil
			}

			if res.Err != nil {
				slog.Error("Error collecting article", "error", res.Err, "pipeline", p.config.Name)
				errorCount++
				continue
			}

			if id, err := p.storer.Save(ctx, res.Result); err != nil {
				slog.Error("Error saving article",
					"error", err,
					"pipeline", p.config.Name,
					"title", res.Result.Title,
				)
				errorCount++
			} else {
				slog.Debug("Article saved successfully",
					"id", id,
					"title", res.Result.Title,
					"pipeline", p.config.Name,
				)
				processedCount++
			}
		}
	}
}

// processBatch handles bulk article processing
func (p *ArticlePipeline) processBatch(ctx context.Context, results <-chan collector.Result[domain.Article]) error {
	var articles []domain.Article
	processedCount := 0
	errorCount := 0
	batchCount := 0

	defer func() {
		if len(articles) > 0 {
			if err := p.storer.SaveBulk(ctx, articles); err != nil {
				slog.Error("Error saving final bulk of articles",
					"error", err,
					"count", len(articles),
					"pipeline", p.config.Name,
				)
			} else {
				slog.Info("Final bulk saved successfully",
					"count", len(articles),
					"pipeline", p.config.Name,
				)
				processedCount += len(articles)
				batchCount++
			}
		}

		slog.Info("Pipeline batch processing completed",
			"pipeline", p.config.Name,
			"total_processed", processedCount,
			"total_errors", errorCount,
			"total_batches", batchCount,
		)
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Pipeline context cancelled, stopping collection",
				"pipeline", p.config.Name,
				"processed", processedCount,
				"errors", errorCount,
				"pending_batch", len(articles),
			)
			return ctx.Err()
		case res, ok := <-results:
			if !ok {
				slog.Info("Collection channel closed, stopping collection",
					"pipeline", p.config.Name,
					"processed", processedCount,
					"errors", errorCount,
					"pending_batch", len(articles),
				)
				return nil
			}

			if res.Err != nil {
				slog.Error("Error collecting article", "error", res.Err, "pipeline", p.config.Name)
				errorCount++
				continue
			}

			articles = append(articles, res.Result)

			if len(articles) >= p.config.Bulk.Size {
				if err := p.storer.SaveBulk(ctx, articles); err != nil {
					slog.Error("Error saving bulk articles",
						"error", err,
						"count", len(articles),
						"pipeline", p.config.Name,
					)
					errorCount += len(articles)
				} else {
					slog.Info("Bulk articles saved successfully",
						"count", len(articles),
						"pipeline", p.config.Name,
						"batch", batchCount+1,
					)
					processedCount += len(articles)
					batchCount++
				}
				articles = articles[:0] // Reset slice
			}
		}
	}
}

// Stop gracefully stops the pipeline
func (p *ArticlePipeline) Stop() {
	slog.Info("Stopping pipeline...", "pipeline", p.config.Name)

	if p.collector != nil {
		// Collector stop logic would go here if available
		slog.Debug("Collector stopped", "pipeline", p.config.Name)
	}

	if p.storer != nil {
		// Storer cleanup logic would go here if available
		p.storer = nil
		slog.Debug("Storer cleaned up", "pipeline", p.config.Name)
	}

	slog.Info("Pipeline stopped", "pipeline", p.config.Name)
}
