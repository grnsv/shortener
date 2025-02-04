package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/models"
)

type Storage interface {
	Save(ctx context.Context, model models.URL) error
	Get(ctx context.Context, short string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}

func NewStorage(ctx context.Context, cfg *config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewDBStorage(ctx, cfg.DatabaseDSN)
	}

	if cfg.FileStoragePath != "" {
		return NewFileStorage(ctx, cfg.FileStoragePath)
	}

	return NewMemoryStorage(ctx)
}

type MemoryStorage struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{urls: make(map[string]string)}, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) Save(ctx context.Context, model models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[model.ShortURL] = model.OriginalURL

	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, short string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	long, exists := s.urls[short]
	if !exists {
		return "", errors.New("not found")
	}

	return long, nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

type FileStorage struct {
	file   *os.File
	writer *bufio.Writer
	memory MemoryStorage
}

func NewFileStorage(ctx context.Context, filename string) (*FileStorage, error) {
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

func (s *FileStorage) Save(ctx context.Context, model models.URL) error {
	if err := json.NewEncoder(s.writer).Encode(model); err != nil {
		return err
	}

	if err := s.writer.Flush(); err != nil {
		return err
	}

	return s.memory.Save(ctx, model)
}

func (s *FileStorage) Get(ctx context.Context, short string) (string, error) {
	return s.memory.Get(ctx, short)
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

type DBStorage struct {
	db DB
}

func NewDBStorage(ctx context.Context, dataSourceName string) (*DBStorage, error) {
	db, err := NewDB(dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = initDB(ctx, db); err != nil {
		return nil, err
	}

	return &DBStorage{db: db}, nil
}

func initDB(ctx context.Context, db DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id uuid NOT NULL,
			short_url text NOT NULL,
			original_url text NOT NULL,
			CONSTRAINT urls_pk PRIMARY KEY (id),
			CONSTRAINT urls_short_url_unique UNIQUE (short_url)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}

func (s *DBStorage) Save(ctx context.Context, model models.URL) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO urls (id, short_url, original_url)
		VALUES ($1::uuid, $2, $3)
		ON CONFLICT DO NOTHING
	`, model.UUID, model.ShortURL, model.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Get(ctx context.Context, short string) (string, error) {
	var long string
	err := s.db.QueryRowContext(ctx, `
		SELECT original_url
		FROM urls
		WHERE short_url = $1
		LIMIT 1
	`, short).Scan(&long)
	if err != nil {
		return "", err
	}

	return long, nil
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
