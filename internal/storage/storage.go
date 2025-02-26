package storage

import (
	"context"
	"errors"

	"github.com/grnsv/shortener/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	ErrAlreadyExist = errors.New("already exist")
	ErrNotFound     = errors.New("not found")
	ErrDeleted      = errors.New("deleted")
)

func New(ctx context.Context, cfg *config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sqlx.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		return NewDBStorage(ctx, &DBWrapper{db})
	}

	if cfg.FileStoragePath != "" {
		return NewFileStorage(ctx, cfg.FileStoragePath)
	}

	return NewMemoryStorage(ctx)
}
