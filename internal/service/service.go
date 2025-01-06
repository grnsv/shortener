package service

import (
	"crypto/md5"
	"encoding/base64"
	"sync"
)

const shortURLLength = 8

type URLStorage interface {
	Save(short, long string)
	Get(short string) (string, bool)
}

type MemoryStorage struct {
	urls map[string]string
	mu   sync.RWMutex
}

func (s *MemoryStorage) Save(short, long string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[short] = long
}

func (s *MemoryStorage) Get(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.urls[short]
	return long, exists
}

type URLShortener struct {
	storage URLStorage
}

func NewURLShortener() *URLShortener {
	return &URLShortener{storage: &MemoryStorage{urls: make(map[string]string)}}
}

func (s *URLShortener) generateShortURL(url string) string {
	h := md5.Sum([]byte(url))
	return base64.URLEncoding.EncodeToString(h[:])[:shortURLLength]
}

func (s *URLShortener) ShortenURL(url string) string {
	shortURL := s.generateShortURL(url)
	s.storage.Save(shortURL, url)
	return shortURL
}

func (s *URLShortener) ExpandURL(shortURL string) (string, bool) {
	return s.storage.Get(shortURL)
}
