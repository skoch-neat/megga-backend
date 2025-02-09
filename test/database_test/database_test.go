package database_test

import (
	"context"
	"megga-backend/internal/config"
	"megga-backend/internal/database"
	"os"
	"testing"

	"github.com/pashagolub/pgxmock"
)

func TestMain(m *testing.M) {
	config.LoadEnv("../../.env")
	exitCode := m.Run() // Run tests
	os.Exit(exitCode)   // Ensure proper exit handling
}

func TestInitDB(t *testing.T) {
	err := database.InitDB()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
		return
	}

	if database.DB == nil {
		t.Fatal("Expected database to be initialized, but it is nil")
	}

	database.CloseDB()
}

func TestCloseDB(t *testing.T) {
	err := database.InitDB()
	if err != nil {
		t.Logf("Failed to initialize database: %v", err)
		return
	}

	if database.DB == nil {
		t.Fatal("Expected initialized database, but got nil")
	}

	database.CloseDB()

	if database.DB != nil {
		t.Fatal("Expected database connection to be closed, but it is still active")
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
