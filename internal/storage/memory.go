package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/grnsv/shortener/internal/models"
)

type MemoryStorage struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{urls: make(map[string]string)}, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) Save(ctx context.Context, model models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[model.ShortURL] = model.OriginalURL

	return nil
}

func (s *MemoryStorage) SaveMany(ctx context.Context, models []models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, model := range models {
		s.urls[model.ShortURL] = model.OriginalURL
	}

	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, short string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.urls[short]
	if !exists {
		return "", errors.New("not found")
	}

	return long, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
