package service

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB interface {
	Close() error
	PingContext(ctx context.Context) error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewDB(dataSourceName string) (DB, error) {
	return sql.Open("pgx", dataSourceName)
}
