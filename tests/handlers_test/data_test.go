package handlers_test

import (
	"bytes"
	"megga-backend/handlers"
	"net/http"
	"net/http/httptest"
	"regexp"
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

	mock.ExpectQuery("SELECT data_id FROM data WHERE series_id =").
		WithArgs("APU0000708111").
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectQuery("INSERT INTO data").
		WithArgs("Eggs, Grade A, Large", "APU0000708111", "per dozen", 4.146, 4.146, "M12", "2024").
		WillReturnRows(pgxmock.NewRows([]string{"data_id"}).AddRow(42))

	body := bytes.NewBufferString(`{
		"series_id": "APU0000708111",
		"latest_value": 4.146,
		"period": "M12",
		"year": "2024"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateData(w, req, mock)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestCreateData_InvalidBody(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	body := bytes.NewBufferString(`{"name": "", "series_id": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateData(w, req, mock)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ✅ Test invalid series_id
func TestCreateData_InvalidSeriesID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	body := bytes.NewBufferString(`{
		"series_id": "",
		"latest_value": 4.146,
		"period": "M12",
		"year": "2024"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateData(w, req, mock)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ✅ Test negative latest_value
func TestCreateData_NegativeValue(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	body := bytes.NewBufferString(`{
		"series_id": "APU0000708111",
		"latest_value": -10.0,
		"period": "M12",
		"year": "2024"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateData(w, req, mock)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetDataByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	lastUpdated := time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT data_id, name, series_id, unit, previous_value, latest_value, last_updated, period, year FROM data WHERE data_id = \$1`).
		WithArgs(42).
		WillReturnRows(pgxmock.NewRows([]string{"data_id", "name", "series_id", "unit", "previous_value", "latest_value", "last_updated", "period", "year"}).
			AddRow(42, "Test Item", "SERIES_123", "kg", 90.0, 100.0, lastUpdated, "M12", "2024"))

	req := httptest.NewRequest(http.MethodGet, "/data/42", nil)
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	handlers.GetDataByID(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateData(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE data SET previous_value = latest_value, latest_value = $1, name = $2, series_id = $3, unit = $4, period = $5, year = $6, last_updated = NOW() WHERE data_id = $7")).
		WithArgs(200.0, "Updated Item", "SERIES_123", "kg", "M12", "2024", 42).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	body := bytes.NewBufferString(`{
		"name": "Updated Item",
		"series_id": "SERIES_123",
		"unit": "kg",
		"latest_value": 200.0,
		"period": "M12",
		"year": "2024"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/data/42", body)
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "42"})

	handlers.UpdateData(w, req, mock)

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

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM data WHERE data_id = $1")).
		WithArgs(99). // ✅ Ensure it matches a non-existent ID
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	req := httptest.NewRequest(http.MethodDelete, "/data/99", nil)
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "99"})

	handlers.DeleteData(w, req, mock)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetDataByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(`SELECT data_id, name, series_id, unit, previous_value, latest_value, last_updated, period, year FROM data WHERE data_id = \$1`).
		WithArgs(99).
		WillReturnError(pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/data/99", nil)
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "99"})

	handlers.GetDataByID(w, req, mock)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteData_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM data").
		WithArgs(99).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	req := httptest.NewRequest(http.MethodDelete, "/data/99", nil)
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "99"})

	handlers.DeleteData(w, req, mock)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRejectNegativeValues(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	body := bytes.NewBufferString(`{
		"series_id": "APU0000708111",
		"latest_value": -10.0,
		"period": "M12",
		"year": "2024"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateData(w, req, mockDB)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
