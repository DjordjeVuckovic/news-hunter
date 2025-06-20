package storage

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/google/uuid"
	"log/slog"
)

type JsonFileStorer struct {
	filePath string
}

func NewJsonFileStorer(filePath string) *JsonFileStorer {
	return &JsonFileStorer{
		filePath: filePath,
	}
}

func (s *JsonFileStorer) Save(ctx context.Context, article domain.Article) (uuid.UUID, error) {
	// Implement the logic to save the article to a JSON file
	// This is a placeholder implementation
	// You would typically use encoding/json to marshal the article and write it to the file
	slog.Info("Saving article to JSON file", "title", article.Title)
	// For now, just return a new UUID
	return uuid.New(), nil
}
