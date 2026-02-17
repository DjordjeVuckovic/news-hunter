package embedding

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Enabled   bool
	Model     string
	MaxLength *int
	BaseURL   string
}

func LoadConfigFromEnv() (*Config, error) {
	enabled := os.Getenv("EMBEDDING_ENABLED")
	model := os.Getenv("EMBEDDING_MODEL")
	maxLen := os.Getenv("EMBEDDING_MAX_LENGTH")
	baseUrl := os.Getenv("EMBEDDING_BASE_URL")

	if baseUrl == "" {
		return nil, errors.New("EMBEDDING_BASE_URL environment variable not set")
	}

	return &Config{
		Enabled: enabled == "true",
		Model:   model,
		MaxLength: func() *int {
			if maxLen == "" {
				return nil
			}
			val, err := strconv.Atoi(maxLen)
			if err != nil {
				return nil
			}
			return &val
		}(),
		BaseURL: baseUrl,
	}, nil
}
