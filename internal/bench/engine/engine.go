package engine

import "github.com/DjordjeVuckovic/news-hunter/internal/storage"

type SearchEngine struct {
	Name     string
	Searcher storage.FtsSearcher
}
