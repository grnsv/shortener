package storage

import (
	"context"
	"database/sql"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
)

type DB interface {
	sqlx.ExtContext
	PingContext(ctx context.Context) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	Close() error
}

type DBStorage struct {
	db DB
}

func NewDBStorage(ctx context.Context, db DB) (*DBStorage, error) {
	storage := &DBStorage{db}
	if err := storage.initDB(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *DBStorage) initDB(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
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

func (s *DBStorage) SaveMany(ctx context.Context, models []models.URL) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO urls (id, short_url, original_url)
        VALUES (:id, :short_url, :original_url)
	`, models)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Get(ctx context.Context, short string) (string, error) {
	var long string
	err := s.db.QueryRowxContext(ctx, `
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
