package storage

import (
	"context"
	"database/sql"
	"io"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
)

type Storage interface {
	Saver
	Retriever
	Deleter
	Pinger
	Closer
}

type Saver interface {
	Save(ctx context.Context, model models.URL) error
	SaveMany(ctx context.Context, models []models.URL) error
}

type Retriever interface {
	Get(ctx context.Context, short string) (string, error)
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
}

type Deleter interface {
	DeleteMany(ctx context.Context, userID string, shortURLs []string) error
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type Closer interface {
	io.Closer
}

type DB interface {
	sqlx.ExtContext
	PingContext(ctx context.Context) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PreparexContext(ctx context.Context, query string) (Stmt, error)
	Closer
}

type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, args ...interface{}) error
	QueryRowxContext(ctx context.Context, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, args ...interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, dest interface{}, args ...interface{}) error
	Closer
}
