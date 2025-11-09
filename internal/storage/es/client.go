package es

import "github.com/elastic/go-elasticsearch/v8"

type ClientConfig struct {
	Addresses []string
	IndexName string
	Username  string
	Password  string
}

func newClient(config ClientConfig) (*elasticsearch.TypedClient, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
	}

	if config.Username != "" && config.Password != "" {
		cfg.Username = config.Username
		cfg.Password = config.Password
	}

	client, err := elasticsearch.NewTypedClient(cfg)

	return client, err
}
