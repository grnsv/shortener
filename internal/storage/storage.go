package storage

import (
	"context"
	"errors"
	"os"

	"github.com/grnsv/shortener/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var ErrAlreadyExist = errors.New("already exist")

func New(ctx context.Context, cfg *config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sqlx.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		return NewDBStorage(ctx, &DBWrapper{db})
	}

	if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}

		return NewFileStorage(ctx, file)
	}

	return NewMemoryStorage(ctx)
}
