package storage

import (
	"context"
	"errors"
	"os"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var ErrAlreadyExist = errors.New("already exist")

type Storage interface {
	Save(ctx context.Context, model models.URL) error
	SaveMany(ctx context.Context, models []models.URL) error
	Get(ctx context.Context, short string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}

func New(ctx context.Context, cfg *config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sqlx.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		return NewDBStorage(ctx, db)
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
