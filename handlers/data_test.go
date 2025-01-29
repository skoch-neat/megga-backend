package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func TestCreateData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	// Verify the correct number of arguments
	mock.ExpectQuery("INSERT INTO data").
		WithArgs("Test Item", "SERIES_123", "good", "kg", 100.0, 100.0, 30).
		WillReturnRows(pgxmock.NewRows([]string{"data_id"}).AddRow(42))

	body := bytes.NewBufferString(`{
		"name": "Test Item",
		"series_id": "SERIES_123",
		"type": "good",
		"unit": "kg",
		"latest_value": 100.0,
		"update_interval_in_days": 30
	}`)
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateData(w, req, mock)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestGetData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	// Mocking `last_updated` as `time.Time`
	lastUpdated := time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT data_id, name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days FROM data").
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "name", "series_id", "type", "unit", "previous_value", "latest_value", "last_updated", "update_interval_in_days"}).
			AddRow(42, "Test Item", "SERIES_123", "good", "kg", 90.0, 100.0, lastUpdated, 30))

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	w := httptest.NewRecorder()

	GetData(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetDataByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	lastUpdated := time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT data_id, name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days FROM data WHERE data_id = \$1`).
		WithArgs(42).
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "name", "series_id", "type", "unit", "previous_value", "latest_value", "last_updated", "update_interval_in_days"}).
			AddRow(42, "Test Item", "SERIES_123", "good", "kg", 90.0, 100.0, lastUpdated, 30))

	req := httptest.NewRequest(http.MethodGet, "/data/42", nil)
	w := httptest.NewRecorder()

	// Inject path variable into request context
	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	GetDataByID(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetDataByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(`SELECT data_id, name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days FROM data WHERE data_id = \$1`).
		WithArgs(99).
		WillReturnError(pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/data/99", nil)
	w := httptest.NewRecorder()

	// Inject path variable into request context
	req = mux.SetURLVars(req, map[string]string{"id": "99"})

	GetDataByID(w, req, mock)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec(`UPDATE data`).
		WithArgs("Updated Item", "SERIES_123", "indicator", "kg", 200.0, 60, 42).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	body := bytes.NewBufferString(`{
		"name": "Updated Item",
		"series_id": "SERIES_123",
		"type": "indicator",
		"unit": "kg",
		"latest_value": 200.0,
		"update_interval_in_days": 60
	}`)
	req := httptest.NewRequest(http.MethodPut, "/data/42", body)
	w := httptest.NewRecorder()

	// Inject path variable into request context
	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	UpdateData(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM data").
		WithArgs(42).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	req := httptest.NewRequest(http.MethodDelete, "/data/42", nil)
	w := httptest.NewRecorder()

	// Inject path variable into request context
	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	DeleteData(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
