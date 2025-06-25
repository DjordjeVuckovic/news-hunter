package ingest

import (
	"context"
	"log/slog"
	"time"

	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
)

const defaultESBatchSize = 500

type EsPipeline struct {
	collector collector.Collector[domain.Article]
	storer    storage.Storer
	config    *PipelineConfig
}

type EsPipelineOption func(pipeline *EsPipeline)

func WithESBulk(size int) EsPipelineOption {
	return func(pipeline *EsPipeline) {
		if pipeline.config.Bulk == nil {
			pipeline.config.Bulk = &BulkOptions{}
		}
		pipeline.config.Bulk.Enabled = true
		pipeline.config.Bulk.Size = size
	}
}

func WithESConfig(config *PipelineConfig) EsPipelineOption {
	return func(pipeline *EsPipeline) {
		pipeline.config = config
	}
}

func NewEsPipeline(c collector.Collector[domain.Article], storer storage.Storer, opts ...EsPipelineOption) *EsPipeline {
	p := &EsPipeline{
		collector: c,
		storer:    storer,
		config: &PipelineConfig{
			Name: "elasticsearch-pipeline",
			Bulk: &BulkOptions{
				Enabled: false,
				Size:    defaultESBatchSize,
			},
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *EsPipeline) Run(ctx context.Context) error {
	start := time.Now()
	slog.Info("Starting EsPipeline run",
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
		runErr = p.importBatch(ctx, results)
	} else {
		runErr = p.importBasic(ctx, results)
	}

	duration := time.Since(start)
	slog.Info("EsPipeline run completed",
		"pipeline", p.config.Name,
		"duration", duration,
		"error", runErr,
	)

	return runErr
}

func (p *EsPipeline) importBasic(ctx context.Context, results <-chan collector.Result[domain.Article]) error {
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
				slog.Error("Error saving article to Elasticsearch",
					"error", err,
					"pipeline", p.config.Name,
					"title", res.Result.Title,
				)
				errorCount++
			} else {
				slog.Debug("Article indexed successfully",
					"id", id,
					"title", res.Result.Title,
					"pipeline", p.config.Name,
				)
				processedCount++
			}
		}
	}
}

func (p *EsPipeline) importBatch(ctx context.Context, results <-chan collector.Result[domain.Article]) error {
	var articles []domain.Article
	processedCount := 0
	errorCount := 0
	batchCount := 0

	defer func() {
		if len(articles) > 0 {
			if err := p.storer.SaveBulk(ctx, articles); err != nil {
				slog.Error("Error saving final bulk of articles to Elasticsearch",
					"error", err,
					"count", len(articles),
					"pipeline", p.config.Name,
				)
			} else {
				slog.Info("Final bulk saved successfully to Elasticsearch",
					"count", len(articles),
					"pipeline", p.config.Name,
				)
				processedCount += len(articles)
				batchCount++
			}
		}

		slog.Info("Elasticsearch pipeline batch processing completed",
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
					slog.Error("Error saving bulk articles to Elasticsearch",
						"error", err,
						"count", len(articles),
						"pipeline", p.config.Name,
					)
					errorCount += len(articles)
				} else {
					slog.Info("Bulk articles saved successfully to Elasticsearch",
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

func (p *EsPipeline) Stop() {
	slog.Info("Stopping Elasticsearch pipeline...", "pipeline", p.config.Name)

	if p.collector != nil {
		// Collector stop logic would go here if available
		slog.Debug("Collector stopped", "pipeline", p.config.Name)
	}

	if p.storer != nil {
		// Storer cleanup logic would go here if available
		p.storer = nil
		slog.Debug("Storer cleaned up", "pipeline", p.config.Name)
	}

	slog.Info("Elasticsearch pipeline stopped", "pipeline", p.config.Name)
}
