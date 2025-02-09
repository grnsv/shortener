package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/grnsv/shortener/internal/models"
)

type File interface {
	io.ReadWriteCloser
}

type FileStorage struct {
	file   File
	writer *bufio.Writer
	memory *MemoryStorage
}

func NewFileStorage(ctx context.Context, file File) (*FileStorage, error) {
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
		memory: &MemoryStorage{urls: urls},
	}, nil
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
