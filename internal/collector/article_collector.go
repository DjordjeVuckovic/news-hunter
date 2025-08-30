package collector

import (
	"context"
	"log/slog"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/reader"
)

type ArticleCollector struct {
	Reader  reader.RawParallelReader
	Mapper  reader.Mapper
	workers int
}

func NewArticleCollector(r reader.RawParallelReader, mapper reader.Mapper) *ArticleCollector {
	return &ArticleCollector{
		Reader:  r,
		Mapper:  mapper,
		workers: defaultWorkers,
	}
}

const defaultWorkers = 10

func (ac *ArticleCollector) Collect(ctx context.Context) (<-chan Result[domain.Article], error) {

	result, err := ac.Reader.ReadParallel(ctx, ac.workers)
	if err != nil {
		return nil, err
	}

	// Create a channel to send the results
	collectionResult := make(chan Result[domain.Article])
	go func() {
		defer close(collectionResult)

		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-result:
				if !ok {
					slog.Info("Reader channel closed, stopping collection")
					return
				}
				if res.Err != nil {
					collectionResult <- Result[domain.Article]{Err: res.Err}
				}

				// Map the record to an Article
				article, err := ac.Mapper.Map(res.Record, nil)
				if err != nil {
					collectionResult <- Result[domain.Article]{Err: err}
					slog.Error("failed to map record to article", "error", err)
					continue
				}

				// Send the mapped article to the channel
				collectionResult <- Result[domain.Article]{Result: article}

			}
		}
	}()

	return collectionResult, nil
}
