package reader

import (
	"github.com/DjordjeVuckovic/news-hunter/pkg/apis/datamapping"
	"gopkg.in/yaml.v3"
	"io"
)

type YAMLConfigLoader struct {
	reader io.Reader
}

func NewYAMLConfigLoader(reader io.Reader) *YAMLConfigLoader {
	return &YAMLConfigLoader{
		reader: reader,
	}
}

func (cl *YAMLConfigLoader) Load(validate bool) (*datamapping.DataMapper, error) {
	decoder := yaml.NewDecoder(cl.reader)
	var mapping datamapping.DataMapper
	if err := decoder.Decode(&mapping); err != nil {
		return nil, err
	}
	if validate {
		if err := mapping.Validate(); err != nil {
			return nil, err
		}
	}
	return &mapping, nil
}
