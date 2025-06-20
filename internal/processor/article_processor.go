package processor

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/collector"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
)

type ArticleProcessor struct {
	collector collector.Collector[domain.Article]
}

func NewArticleProcessor(c collector.Collector[domain.Article]) *ArticleProcessor {
	return &ArticleProcessor{
		collector: c,
	}
}
