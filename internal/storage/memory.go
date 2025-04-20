package storage

import (
	"context"
	"sync"

	"github.com/grnsv/shortener/internal/models"
)

// MemoryStorage implements an in-memory storage for URL mappings.
// It is safe for concurrent use and is primarily intended for development or testing environments.
// The storage uses a sync.Map to store short URL to original URL mappings.
type MemoryStorage struct {
	urls sync.Map
}

// NewMemoryStorage creates and returns a new in-memory storage instance.
func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{}, nil
}

// Close closes the in-memory storage. It is a no-op for MemoryStorage.
func (s *MemoryStorage) Close() error {
	return nil
}

// Save stores a single URL mapping in memory.
func (s *MemoryStorage) Save(ctx context.Context, model models.URL) error {
	s.urls.Store(model.ShortURL, model.OriginalURL)
	return nil
}

// SaveMany stores multiple URL mappings in memory.
func (s *MemoryStorage) SaveMany(ctx context.Context, models []models.URL) error {
	for _, model := range models {
		s.urls.Store(model.ShortURL, model.OriginalURL)
	}
	return nil
}

// Get retrieves the original URL for a given short URL from memory.
func (s *MemoryStorage) Get(ctx context.Context, short string) (string, error) {
	value, ok := s.urls.Load(short)
	if !ok {
		return "", ErrNotFound
	}
	return value.(string), nil
}

// Ping checks the availability of the in-memory storage. Always returns nil.
func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

// GetAll returns all URL mappings for a user from memory.
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

// DeleteMany deletes multiple short URLs for a user from memory.
func (s *MemoryStorage) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	for _, shortURL := range shortURLs {
		s.urls.Delete(shortURL)
	}

	return nil
}
