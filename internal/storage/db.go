package storage

import (
	"context"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// DBWrapper wraps a sqlx.DB to implement the DB interface.
type DBWrapper struct {
	*sqlx.DB
}

// PreparexContext prepares a statement with context and returns a wrapped statement.
func (db *DBWrapper) PreparexContext(ctx context.Context, query string) (Stmt, error) {
	stmt, err := db.DB.PreparexContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &StmtWrapper{stmt}, nil
}

// StmtWrapper wraps a sqlx.Stmt to implement the Stmt interface.
type StmtWrapper struct {
	*sqlx.Stmt
}

// DBStorage provides methods to interact with the URLs database.
type DBStorage struct {
	db           DB
	saveStmt     Stmt
	getAllStmt   Stmt
	getStmt      Stmt
	deleteStmt   Stmt
	getStatsStmt Stmt
}

// NewDBStorage creates a new DBStorage and initializes the database schema and prepared statements.
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

	if s.getStatsStmt, err = s.db.PreparexContext(ctx, `
		SELECT
			COUNT(*) AS urls_count,
			COUNT(DISTINCT user_id) AS users_count
		FROM
			urls;
	`); err != nil {
		return err
	}

	return nil
}

// Close closes all prepared statements and the underlying database connection.
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
	if err := s.getStatsStmt.Close(); err != nil {
		return err
	}
	if err := s.db.Close(); err != nil {
		return err
	}
	return nil
}

// Save inserts a new URL record into the database.
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

// SaveMany inserts multiple URL records into the database.
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

// Get retrieves the original URL for a given short URL.
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

// Ping checks the database connection.
func (s *DBStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// GetAll retrieves all URLs for a given user.
func (s *DBStorage) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	var urls []models.URL
	if err := s.getAllStmt.SelectContext(ctx, &urls, userID); err != nil {
		return nil, err
	}

	return urls, nil
}

// DeleteMany marks multiple URLs as deleted for a given user.
func (s *DBStorage) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	_, err := s.deleteStmt.ExecContext(ctx, userID, pq.Array(shortURLs))
	return err
}

// GetStats retrieves service statistics and populates the provided Stats struct.
func (s *DBStorage) GetStats(ctx context.Context, stats *models.Stats) error {
	return s.getStatsStmt.GetContext(ctx, stats)
}
