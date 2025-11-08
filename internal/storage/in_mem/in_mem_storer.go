package in_mem

import (
	"context"
	"log/slog"
	"sync"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/google/uuid"
)

type InMemStorer struct {
	storageLock sync.RWMutex
	storage     map[uuid.UUID]domain.Article
}

func NewInMemStorer() *InMemStorer {
	return &InMemStorer{
		storage: make(map[uuid.UUID]domain.Article),
	}
}

func (s *InMemStorer) Save(ctx context.Context, article domain.Article) (uuid.UUID, error) {
	// Implement the logic to save the article to a JSON file
	// This is a placeholder implementation
	// You would typically use encoding/json to marshal the article and write it to the file
	slog.Info("Saving article to JSON file", "title", article.Title)
	s.storageLock.Lock()
	defer s.storageLock.Unlock()
	s.storage[article.ID] = article
	// For now, just return a new UUID
	return uuid.New(), nil
}

func (s *InMemStorer) SaveBulk(ctx context.Context, articles []domain.Article) error {
	s.storageLock.Lock()
	defer s.storageLock.Unlock()

	for _, article := range articles {
		if article.ID == uuid.Nil {
			article.ID = uuid.New()
		}
		s.storage[article.ID] = article
		slog.Info("Saving article to in-memory storage", "title", article.Title, "id", article.ID)
	}

	return nil
}
