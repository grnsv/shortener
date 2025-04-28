// Package storage provides interfaces and implementations for persistent and in-memory
// storage backends used by the URL shortener service. It supports PostgreSQL, file-based,
// and in-memory storage options, and defines common error variables for storage operations.
package storage

import (
	"context"
	"errors"

	"github.com/grnsv/shortener/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Storage error variables used throughout the storage package.
var (
	ErrAlreadyExist = errors.New("already exist")
	ErrNotFound     = errors.New("not found")
	ErrDeleted      = errors.New("deleted")
)

// New creates a new Storage implementation based on the provided configuration.
// It selects the storage backend in the following order: PostgreSQL, file storage, or in-memory storage.
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
