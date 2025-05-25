package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/grnsv/shortener/internal/models"
)

// FileStorage implements persistent storage using a file and in-memory cache.
type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory *MemoryStorage
}

// NewFileStorage creates a new FileStorage instance with the given file path.
func NewFileStorage(ctx context.Context, path string) (*FileStorage, error) {
	file, writer, err := openFile(path)
	if err != nil {
		return nil, err
	}
	memory, err := NewMemoryStorage(ctx)
	if err != nil {
		return nil, err
	}

	storage := &FileStorage{
		file:   file,
		writer: writer,
		memory: memory,
	}
	if err = storage.loadFromFile(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func openFile(path string) (*os.File, *bufio.Writer, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}

	return file, bufio.NewWriter(file), nil
}

func (s *FileStorage) loadFromFile(ctx context.Context) error {
	var err error
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		model := &models.URL{}
		if err = json.Unmarshal(scanner.Bytes(), model); err != nil {
			return err
		}

		if err = s.memory.Save(ctx, *model); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying file and memory storage.
func (s *FileStorage) Close() error {
	if err := s.writer.Flush(); err != nil {
		return err
	}
	if err := s.memory.Close(); err != nil {
		return err
	}
	return s.file.Close()
}

// Save persists a single URL model to file and memory.
func (s *FileStorage) Save(ctx context.Context, model models.URL) error {
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}
	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.Save(ctx, model)
}

// SaveMany persists multiple URL models to file and memory.
func (s *FileStorage) SaveMany(ctx context.Context, models []models.URL) error {
	encoder := json.NewEncoder(s.writer)
	for _, model := range models {
		if err := encoder.Encode(model); err != nil {
			return err
		}
	}
	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.SaveMany(ctx, models)
}

// Get retrieves the original URL for a given short URL from memory.
func (s *FileStorage) Get(ctx context.Context, short string) (string, error) {
	return s.memory.Get(ctx, short)
}

// Ping checks the availability of the storage (always returns nil).
func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

// GetAll returns all URL models for a given user from memory.
func (s *FileStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	return s.memory.GetAll(ctx, userID)
}

// DeleteMany deletes multiple short URLs for a user and updates the file.
func (s *FileStorage) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	if err := s.memory.DeleteMany(ctx, userID, shortURLs); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "storage")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(tempFile)

	encoder := json.NewEncoder(writer)
	s.memory.urls.Range(func(key, value interface{}) bool {
		if err = encoder.Encode(value.(models.URL)); err != nil {
			return false
		}
		return true
	})

	if err = writer.Flush(); err != nil {
		return err
	}
	if err = tempFile.Close(); err != nil {
		return err
	}
	if err = s.file.Close(); err != nil {
		return err
	}
	if err = os.Rename(tempFile.Name(), s.file.Name()); err != nil {
		return err
	}
	s.file, s.writer, err = openFile(s.file.Name())
	if err != nil {
		return err
	}

	return nil
}

// GetStats retrieves service statistics and populates the provided Stats struct.
func (s *FileStorage) GetStats(ctx context.Context, stats *models.Stats) error {
	return s.memory.GetStats(ctx, stats)
}
