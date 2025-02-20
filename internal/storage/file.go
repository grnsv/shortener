package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/grnsv/shortener/internal/models"
)

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory *MemoryStorage
}

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

func (s *FileStorage) Close() error {
	if err := s.memory.Close(); err != nil {
		return err
	}
	return s.file.Close()
}

func (s *FileStorage) Save(ctx context.Context, model models.URL) error {
	defer s.writer.Flush()
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}

	return s.memory.Save(ctx, model)
}

func (s *FileStorage) SaveMany(ctx context.Context, models []models.URL) error {
	defer s.writer.Flush()
	encoder := json.NewEncoder(s.writer)
	for _, model := range models {
		if err := encoder.Encode(model); err != nil {
			return err
		}
	}

	return s.memory.SaveMany(ctx, models)
}

func (s *FileStorage) Get(ctx context.Context, short string) (string, error) {
	return s.memory.Get(ctx, short)
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *FileStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	return s.memory.GetAll(ctx, userID)
}

func (s *FileStorage) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	if err := s.memory.DeleteMany(ctx, userID, shortURLs); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "storage")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	writer := bufio.NewWriter(tempFile)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	s.memory.urls.Range(func(key, value interface{}) bool {
		shortURL, ok1 := key.(string)
		originalURL, ok2 := value.(string)
		if ok1 && ok2 {
			model := models.URL{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			}
			if err := encoder.Encode(model); err != nil {
				return false
			}
		}
		return true
	})

	if err := s.file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempFile.Name(), s.file.Name()); err != nil {
		return err
	}
	s.file, s.writer, err = openFile(s.file.Name())
	if err != nil {
		return err
	}

	return nil
}
