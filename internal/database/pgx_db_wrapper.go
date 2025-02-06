package database

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PgxDBWrapper struct {
	Pool *pgxpool.Pool
}

func (w *PgxDBWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return w.Pool.Exec(ctx, sql, arguments...)
}

func (w *PgxDBWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return w.Pool.QueryRow(ctx, sql, args...)
}

func (w *PgxDBWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return w.Pool.Query(ctx, sql, args...)
}
