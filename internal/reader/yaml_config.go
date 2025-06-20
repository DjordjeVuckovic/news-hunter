package reader

import (
	"github.com/DjordjeVuckovic/news-hunter/pkg/apis"
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

func (cl *YAMLConfigLoader) Load(validate bool) (*apis.DataMapping, error) {
	decoder := yaml.NewDecoder(cl.reader)
	var mapping apis.DataMapping
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
