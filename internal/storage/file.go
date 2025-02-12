package storage

import (
	"bufio"
	"context"
	"encoding/json"

	"github.com/grnsv/shortener/internal/models"
)

type FileStorage struct {
	file   File
	writer *bufio.Writer
	memory *MemoryStorage
}

func NewFileStorage(ctx context.Context, file File) (*FileStorage, error) {
	writer := bufio.NewWriter(file)
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
	for _, model := range models {
		if err := json.NewEncoder(s.writer).Encode(model); err != nil {
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
