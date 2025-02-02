package service

import (
	"context"
	"database/sql"

	"github.com/grnsv/shortener/internal/config"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB interface {
	Close() error
	PingContext(ctx context.Context) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewDB() (DB, error) {
	return sql.Open("pgx", config.Get().DatabaseDSN)
}
