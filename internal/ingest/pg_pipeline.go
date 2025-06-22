package ingest

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"log/slog"
	"time"
)

const defaultBatchSize = 1000

type PipelineBulkOptions struct {
	enabled bool
	Size    int
}

type PgPipeline struct {
	collector collector.Collector[domain.Article]
	storer    storage.Storer
	bulk      *PipelineBulkOptions
}

type PgPipelineOption func(pipeline *PgPipeline)

func WithPgBulk(size int) PgPipelineOption {
	return func(pipeline *PgPipeline) {
		pipeline.bulk = &PipelineBulkOptions{
			enabled: true,
			Size:    size,
		}
	}
}

func NewPgPipeline(c collector.Collector[domain.Article], storer storage.Storer, opts ...PgPipelineOption) *PgPipeline {
	p := &PgPipeline{
		collector: c,
		storer:    storer,
		bulk: &PipelineBulkOptions{
			enabled: false,
			Size:    defaultBatchSize,
		},
	}
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *PgPipeline) Run(ctx context.Context) error {
	start := time.Now() // Start timer

	results, err := p.collector.Collect(ctx)
	if err != nil {
		slog.Error("Error collecting articles", "error", err)
	}

	var runErr error
	if p.bulk.enabled {
		runErr = p.importBatch(ctx, results)
	} else {
		runErr = p.importBasic(ctx, results)
	}

	duration := time.Since(start)
	slog.Info("PgPipeline run completed", "duration", duration)

	return runErr
}

func (p *PgPipeline) importBasic(ctx context.Context, results <-chan collector.Result[domain.Article]) error {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Pipeline context cancelled, stopping collection")
			return ctx.Err()
		case res, ok := <-results:
			if !ok {
				slog.Info("Collection channel closed, stopping collection")
				return nil
			}
			if res.Err != nil {
				slog.Error("Error collecting article", "error", res.Err)
				continue
			}

			if id, err := p.storer.Save(ctx, res.Result); err != nil {
				slog.Error("Error saving article", "error", err)
			} else {
				slog.Info("Article saved successfully", "id", id, "title", res.Result.Title)
			}
		}
	}
}

func (p *PgPipeline) importBatch(ctx context.Context, results <-chan collector.Result[domain.Article]) error {
	var articles []domain.Article
	defer func() {
		if len(articles) > 0 {
			if err := p.storer.SaveBulk(ctx, articles); err != nil {
				slog.Error("Error saving final bulk of articles", "error", err, "count", len(articles))
			} else {
				slog.Info("Final bulk saved successfully", "count", len(articles))
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Pipeline context cancelled, stopping collection")
			return ctx.Err()
		case res, ok := <-results:
			if !ok {
				slog.Info("Collection channel closed, stopping collection")
				return nil
			}
			if res.Err != nil {
				slog.Error("Error collecting article", "error", res.Err)
				continue
			}

			articles = append(articles, res.Result)

			if len(articles) >= p.bulk.Size {
				if err := p.storer.SaveBulk(ctx, articles); err != nil {
					slog.Error("Error saving bulk articles", "error", err, "count", len(articles))
				} else {
					slog.Info("Bulk articles saved successfully", "count", len(articles))
				}
				articles = articles[:0]
			}
		}
	}
}

func (p *PgPipeline) Stop() {
	slog.Info("Stopping pipeline...")
	if p.collector != nil {
		// p.collector.Stop()
	}
	if p.storer != nil {
		p.storer = nil // Assuming storer has a Stop method, otherwise this is just a placeholder
	}
	slog.Info("Pipeline stopped")
}
