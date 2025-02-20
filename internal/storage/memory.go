package storage

import (
	"context"
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
		return "", ErrNotFound
	}
	return value.(string), nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	var urls []models.URL

	s.urls.Range(func(key, value interface{}) bool {
		shortURL, ok1 := key.(string)
		originalURL, ok2 := value.(string)
		if ok1 && ok2 {
			urls = append(urls, models.URL{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			})
		}
		return true
	})

	return urls, nil
}

func (s *MemoryStorage) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	for _, shortURL := range shortURLs {
		s.urls.Delete(shortURL)
	}

	return nil
}
