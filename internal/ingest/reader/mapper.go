package reader

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/types/document"
	"github.com/DjordjeVuckovic/news-hunter/pkg/apis/datamapping"
)

type MappingOptions struct{}

type Mapper interface {
	Map(map[string]string, *MappingOptions) (document.Article, error)
}

type MappingLoader interface {
	Load(validate bool) (*datamapping.DataMapper, error)
}
