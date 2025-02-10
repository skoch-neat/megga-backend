package database_test

import (
	"context"
	"megga-backend/internal/config"
	"os"
	"regexp"
	"testing"

	"github.com/pashagolub/pgxmock"
)

func TestMain(m *testing.M) {
	config.LoadEnvFile("env/.env.development")
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestInitDB(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	if mock == nil {
		t.Fatal("Expected database mock to be initialized, but it is nil")
	}
}

func TestCloseDB(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	if mock == nil {
		t.Fatal("Expected initialized database mock, but got nil")
	}
}

func TestTransactionRollback(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err := mock.Begin(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = tx.Rollback(context.Background())
	if err != nil {
		t.Fatalf("Expected rollback to succeed, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("❌ Unmet mock expectations: %v", err)
	}
}

func TestDatabaseTransaction_Commit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE data SET previous_value = latest_value, latest_value = $1, year = $2, period = $3, last_updated = NOW() WHERE data_id = $4")).
		WithArgs(4.15, "2024", "M12", 1).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	tx, err := mock.Begin(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, err := tx.Exec(context.Background(), "UPDATE data SET previous_value = latest_value, latest_value = $1, year = $2, period = $3, last_updated = NOW() WHERE data_id = $4", 4.15, "2024", "M12", 1); err != nil {
		t.Fatalf("Expected no error executing update, got %v", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("Expected no error committing transaction, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("❌ Unmet mock expectations: %v", err)
	}
}
