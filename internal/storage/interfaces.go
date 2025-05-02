package storage

import (
	"context"
	"database/sql"
	"io"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
)

//go:generate go tool mockgen -destination=../mocks/mock_storage.go -package=mocks github.com/grnsv/shortener/internal/storage Storage,DB,Stmt

// Storage is the main interface that combines Saver, Retriever, Deleter, Pinger, and Closer interfaces.
type Storage interface {
	Saver
	Retriever
	Deleter
	Pinger
	Closer
}

// Saver provides methods for saving URL models.
type Saver interface {
	Save(ctx context.Context, model models.URL) error
	SaveMany(ctx context.Context, models []models.URL) error
}

// Retriever provides methods for retrieving URL models.
type Retriever interface {
	Get(ctx context.Context, short string) (string, error)
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
}

// Deleter provides a method for deleting multiple short URLs for a user.
type Deleter interface {
	DeleteMany(ctx context.Context, userID string, shortURLs []string) error
}

// Pinger provides a method to check the health of the storage.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Closer provides a method to close the storage and release resources.
type Closer interface {
	io.Closer
}

// DB abstracts a database connection with extended context support.
type DB interface {
	sqlx.ExtContext
	PingContext(ctx context.Context) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PreparexContext(ctx context.Context, query string) (Stmt, error)
	Closer
}

// Stmt abstracts a prepared statement with extended context support.
type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, args ...interface{}) error
	QueryRowxContext(ctx context.Context, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, args ...interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, dest interface{}, args ...interface{}) error
	Closer
}
