package reader

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/pkg/apis/datamapping"
)

type ArticleMapper struct {
	cfg *datamapping.DataMapper
}

func NewArticleMapper(cfg *datamapping.DataMapper) *ArticleMapper {
	return &ArticleMapper{
		cfg: cfg,
	}
}

func (m *ArticleMapper) Map(record map[string]string, _ *MappingOptions) (domain.Article, error) {
	if err := m.cfg.Validate(); err != nil {
		return domain.Article{}, err
	}

	article := domain.Article{}
	val := reflect.ValueOf(&article).Elem()

	for _, fm := range m.cfg.FieldMappings {
		sourceVal := record[fm.Source]

		if sourceVal == "" && !fm.Required {
			slog.Debug("skipping empty field", "field", fm.Source)
			continue
		}

		path := strings.Split(fm.Target, ".")

		if len(path) > 1 {
			err := SetNestedField(val, path, sourceVal, fm.SourceType, m.cfg.DateFormat)
			if err != nil {
				if fm.Required {
					slog.Error("failed to set nested field", "field", fm.Target, "error", err)
					return domain.Article{}, err
				} else {
					slog.Warn("skipping optional nested field", "field", fm.Target, "error", err)
					continue
				}
			}

			continue
		}

		err := SetFlatField(val, path[0], sourceVal, fm.SourceType, m.cfg.DateFormat)
		if err != nil {
			if fm.Required {
				slog.Error("failed to set flat field", "field", fm.Target, "error", err)
				return domain.Article{}, err
			} else {
				slog.Warn("skipping optional field", "field", fm.Target, "error", err)
				continue
			}
		}
	}
	return article, nil
}
