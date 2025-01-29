package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pashagolub/pgxmock"
)

// TestCreateThreshold ensures a threshold can be created
func TestCreateThreshold(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("INSERT INTO thresholds").
		WithArgs(1, 100.5). // Assuming data_id = 1, threshold_value = 100.5
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

// TestGetThresholds ensures all thresholds can be retrieved
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

	GetThresholds(w, req, mock)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
