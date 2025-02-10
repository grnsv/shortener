package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/grnsv/shortener/internal/models"
)

type MemoryStorage struct {
	urls sync.Map
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{}, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) Save(ctx context.Context, model models.URL) error {
	s.urls.Store(model.ShortURL, model.OriginalURL)
	return nil
}

func (s *MemoryStorage) SaveMany(ctx context.Context, models []models.URL) error {
	for _, model := range models {
		s.urls.Store(model.ShortURL, model.OriginalURL)
	}
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, short string) (string, error) {
	value, ok := s.urls.Load(short)
	if !ok {
		return "", errors.New("not found")
	}
	return value.(string), nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
