package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/models"
)

type Storage interface {
	Save(model models.URL) error
	Get(short string) (string, bool)
	Close() error
}

func NewStorage(typ string) (Storage, error) {
	switch typ {
	case "memory":
		return NewMemoryStorage()
	case "file":
		return NewFileStorage(config.Get().FileStoragePath)
	default:
		return nil, fmt.Errorf("unknown type: %s", typ)
	}
}

type MemoryStorage struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{urls: make(map[string]string)}, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) Save(model models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[model.ShortURL] = model.OriginalURL

	return nil
}

func (s *MemoryStorage) Get(short string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.urls[short]
	return long, exists
}

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory MemoryStorage
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	urls := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		model := &models.URL{}
		if err := json.Unmarshal(scanner.Bytes(), model); err != nil {
			return nil, err
		}
		urls[model.ShortURL] = model.OriginalURL
	}

	return &FileStorage{
		file:   file,
		writer: bufio.NewWriter(file),
		memory: MemoryStorage{urls: urls},
	}, nil
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func (s *FileStorage) Save(model models.URL) error {
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.Save(model)
}

func (s *FileStorage) Get(short string) (string, bool) {
	return s.memory.Get(short)
}
