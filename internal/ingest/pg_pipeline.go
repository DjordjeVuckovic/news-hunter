package ingest

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"log/slog"
)

type PgPipeline struct {
	collector collector.Collector[domain.Article]
	storer    storage.Storer
}

func NewPipeline(c collector.Collector[domain.Article], storer storage.Storer) *PgPipeline {
	return &PgPipeline{
		collector: c,
		storer:    storer,
	}
}

func (p *PgPipeline) Run(ctx context.Context) error {
	results, err := p.collector.Collect(ctx)
	if err != nil {
		slog.Error("Error collecting articles", "error", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				slog.Info("Pipeline context cancelled, stopping collection")
				return
			case res, ok := <-results:
				if !ok {
					slog.Info("Collection channel closed, stopping collection")
					return
				}
				if res.Err != nil {
					slog.Error("Error collecting article", "error", res.Err)
				}

				save, err := p.storer.Save(ctx, res.Result)
				if err != nil {
					slog.Error("Error saving article", "error", err, "article", res.Result)
				}
				slog.Info("Article saved successfully", "id", save, "title", res.Result.Title)
			}
		}
	}()
	<-done

	return nil
}
