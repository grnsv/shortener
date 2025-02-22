package storage

import (
	"context"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type DBWrapper struct {
	*sqlx.DB
}

func (db *DBWrapper) PreparexContext(ctx context.Context, query string) (Stmt, error) {
	stmt, err := db.DB.PreparexContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &StmtWrapper{stmt}, nil
}

type StmtWrapper struct {
	*sqlx.Stmt
}

type DBStorage struct {
	db         DB
	saveStmt   Stmt
	getAllStmt Stmt
	getStmt    Stmt
	deleteStmt Stmt
}

func NewDBStorage(ctx context.Context, db DB) (*DBStorage, error) {
	storage := &DBStorage{db: db}
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
			user_id uuid NOT NULL,
			is_deleted boolean NOT NULL DEFAULT false,
			CONSTRAINT urls_pk PRIMARY KEY (id),
			CONSTRAINT urls_short_url_unique UNIQUE (short_url)
		);
		CREATE INDEX IF NOT EXISTS urls_user_id_idx ON urls (user_id);
	`)
	if err != nil {
		return err
	}

	if s.saveStmt, err = s.db.PreparexContext(ctx, `
		INSERT INTO urls (id, user_id, short_url, original_url)
		VALUES ($1::uuid, $2::uuid, $3, $4)
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	if s.getAllStmt, err = s.db.PreparexContext(ctx, `
		SELECT
			short_url,
			original_url
		FROM
			urls
		WHERE
			user_id = $1::uuid
		LIMIT $2 OFFSET $3
	`); err != nil {
		return err
	}

	if s.getStmt, err = s.db.PreparexContext(ctx, `
		SELECT *
		FROM urls
		WHERE short_url = $1
		LIMIT 1
	`); err != nil {
		return err
	}

	if s.deleteStmt, err = s.db.PreparexContext(ctx, `
		UPDATE urls
		SET is_deleted = true
		WHERE user_id = $1 AND short_url = ANY($2)
	`); err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Close() error {
	if err := s.getStmt.Close(); err != nil {
		return err
	}
	if err := s.getAllStmt.Close(); err != nil {
		return err
	}
	if err := s.saveStmt.Close(); err != nil {
		return err
	}
	if err := s.deleteStmt.Close(); err != nil {
		return err
	}
	if err := s.db.Close(); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) Save(ctx context.Context, model models.URL) error {
	result, err := s.saveStmt.ExecContext(ctx, model.UUID, model.UserID, model.ShortURL, model.OriginalURL)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAlreadyExist
	}

	return nil
}

func (s *DBStorage) SaveMany(ctx context.Context, models []models.URL) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO urls (id, user_id, short_url, original_url)
		VALUES (:id, :user_id, :short_url, :original_url)
	`, models)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) Get(ctx context.Context, short string) (string, error) {
	var url models.URL
	err := s.getStmt.GetContext(ctx, &url, short)
	if err != nil {
		return "", err
	}
	if url.IsDeleted {
		return "", ErrDeleted
	}

	return url.OriginalURL, nil
}

func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *DBStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	var allUrls []models.URL
	chunkSize := 1000
	offset := 0

	for {
		var urls []models.URL
		if err := s.getAllStmt.SelectContext(ctx, &urls, userID, chunkSize, offset); err != nil {
			return nil, err
		}

		if len(urls) == 0 {
			break
		}

		allUrls = append(allUrls, urls...)
		offset += chunkSize
	}

	return allUrls, nil
}

func (s *DBStorage) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	_, err := s.deleteStmt.ExecContext(ctx, userID, pq.Array(shortURLs))
	return err
}
