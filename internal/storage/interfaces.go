package storage

import (
	"context"
	"database/sql"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
)

type Storage interface {
	Save(ctx context.Context, model models.URL) error
	SaveMany(ctx context.Context, models []models.URL) error
	Get(ctx context.Context, short string) (string, error)
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
	DeleteMany(ctx context.Context, userID string, shortURLs []string) error
	Ping(ctx context.Context) error
	Close() error
}

type DB interface {
	sqlx.ExtContext
	PingContext(ctx context.Context) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PreparexContext(ctx context.Context, query string) (Stmt, error)
	Close() error
}

type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, args ...interface{}) error
	QueryRowxContext(ctx context.Context, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, args ...interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, dest interface{}, args ...interface{}) error
	Close() error
}
