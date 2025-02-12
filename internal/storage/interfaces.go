package storage

import (
	"context"
	"database/sql"
	"io"

	"github.com/grnsv/shortener/internal/models"
	"github.com/jmoiron/sqlx"
)

type Storage interface {
	Save(ctx context.Context, model models.URL) error
	SaveMany(ctx context.Context, models []models.URL) error
	Get(ctx context.Context, short string) (string, error)
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
	Close() error
}

type File interface {
	io.ReadWriteCloser
}
