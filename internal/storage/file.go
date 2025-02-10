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
	urls, err := loadFromFile(file)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(file)
	storage := &FileStorage{
		file:   file,
		writer: writer,
		memory: &MemoryStorage{urls: urls},
	}
	defer writer.Flush()
	return storage, nil
}

func loadFromFile(file File) (map[string]string, error) {
	urls := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		model := &models.URL{}
		if err := json.Unmarshal(scanner.Bytes(), model); err != nil {
			return nil, err
		}
		urls[model.ShortURL] = model.OriginalURL
	}

	return urls, nil
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func (s *FileStorage) Save(ctx context.Context, model models.URL) error {
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.Save(ctx, model)
}

func (s *FileStorage) SaveMany(ctx context.Context, models []models.URL) error {
	for _, model := range models {
		if err := json.NewEncoder(s.writer).Encode(model); err != nil {
			return err
		}
	}

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.SaveMany(ctx, models)
}

func (s *FileStorage) Get(ctx context.Context, short string) (string, error) {
	return s.memory.Get(ctx, short)
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}
