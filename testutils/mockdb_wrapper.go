package testutils

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

type MockDBWrapper struct {
	Mock pgxmock.PgxPoolIface
}

// NewMockDB creates a new mock database connection
func NewMockDB() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic("Failed to create mock database: " + err.Error())
	}
	return db, mock
}

// MockRow is a helper for creating rows in tests
func MockRow(columns []string, values ...driver.Value) *sqlmock.Rows {
	return sqlmock.NewRows(columns).AddRow(values...)
}

func (m *MockDBWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return m.Mock.Exec(ctx, sql, arguments...)
}

func (m *MockDBWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.Mock.QueryRow(ctx, sql, args...)
}

func (m *MockDBWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return m.Mock.Query(ctx, sql, args...)
}
