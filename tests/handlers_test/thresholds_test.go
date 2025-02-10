package handlers_test

import (
	"errors"
	"megga-backend/handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pashagolub/pgxmock"
)

func TestDeleteThreshold_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM thresholds").
		WithArgs(42).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	req := httptest.NewRequest(http.MethodDelete, "/thresholds/42", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	w := httptest.NewRecorder()

	handlers.DeleteThreshold(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteThreshold_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM thresholds").
		WithArgs(99).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	req := httptest.NewRequest(http.MethodDelete, "/thresholds/99", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99"})
	w := httptest.NewRecorder()

	handlers.DeleteThreshold(w, req, mock)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDeleteThreshold_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM thresholds").
		WithArgs(42).
		WillReturnError(errors.New("database error"))

	req := httptest.NewRequest(http.MethodDelete, "/thresholds/42", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	w := httptest.NewRecorder()

	handlers.DeleteThreshold(w, req, mock)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
