package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pashagolub/pgxmock"
)

func TestCreateThreshold(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("INSERT INTO thresholds").
		WithArgs(1, 100.5).
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id"}).AddRow(42))

	body := bytes.NewBufferString(`{"data_id":1, "threshold_value":100.5}`)
	req := httptest.NewRequest(http.MethodPost, "/thresholds", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateThreshold(w, req, mock)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestGetThresholds(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT threshold_id, data_id, threshold_value FROM thresholds").
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "data_id", "threshold_value"}).
			AddRow(42, 1, 100.5))

	req := httptest.NewRequest(http.MethodGet, "/thresholds", nil)
	w := httptest.NewRecorder()

	GetThresholdsForUser(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateThreshold(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("UPDATE thresholds").
		WithArgs(200.0, 42).
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "data_id", "threshold_value"}).
			AddRow(42, 1, 200.0))

	body := bytes.NewBufferString(`{"threshold_value":200.0}`)
	req := httptest.NewRequest(http.MethodPut, "/thresholds/42", body)
	req.Header.Set("Content-Type", "application/json")

	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	w := httptest.NewRecorder()
	UpdateThreshold(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

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

	DeleteThreshold(w, req, mock)

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

	DeleteThreshold(w, req, mock)

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

	DeleteThreshold(w, req, mock)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
