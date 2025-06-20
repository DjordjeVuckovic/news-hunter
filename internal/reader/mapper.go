package reader

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/pkg/apis"
)

type MappingOptions struct {
	strict bool
}

type Mapper interface {
	Map(map[string]string, *MappingOptions) (domain.Article, error)
}

type MappingLoader interface {
	Load(validate bool) (*apis.DataMapping, error)
}
