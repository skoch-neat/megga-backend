package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

func setupThresholdRecipientRouter(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	RegisterThresholdRecipientRoutes(router, mock)
	return router
}

func TestCreateThresholdRecipient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("INSERT INTO threshold_recipients").
		WithArgs(1, 2, true).
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "recipient_id", "is_user"}).AddRow(1, 2, true))

	router := setupThresholdRecipientRouter(mock)

	body := bytes.NewBufferString(`{
		"threshold_id": 1,
		"recipient_id": 2,
		"is_user": true
	}`)
	req := httptest.NewRequest(http.MethodPost, "/threshold_recipients", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestGetThresholdRecipients(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT threshold_id, recipient_id, is_user FROM threshold_recipients").
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "recipient_id", "is_user"}).
			AddRow(1, 2, true))

	router := setupThresholdRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/threshold_recipients", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetThresholdRecipientByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT threshold_id, recipient_id, is_user FROM threshold_recipients WHERE threshold_id = \\$1 AND recipient_id = \\$2").
		WithArgs(1, 2).
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "recipient_id", "is_user"}).
			AddRow(1, 2, true))

	router := setupThresholdRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/threshold_recipients/1/2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetThresholdRecipientByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT threshold_id, recipient_id, is_user FROM threshold_recipients WHERE threshold_id = \\$1 AND recipient_id = \\$2").
    WithArgs(99, 100).
    WillReturnError(pgx.ErrNoRows)

	router := setupThresholdRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/threshold_recipients/99/100", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateThresholdRecipient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("UPDATE threshold_recipients").
		WithArgs(false, 1, 2).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	router := setupThresholdRecipientRouter(mock)

	body := bytes.NewBufferString(`{
		"is_user": false
	}`)
	req := httptest.NewRequest(http.MethodPut, "/threshold_recipients/1/2", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteThresholdRecipient_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM threshold_recipients").
		WithArgs(1, 2).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	router := setupThresholdRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodDelete, "/threshold_recipients/1/2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteThresholdRecipient_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM threshold_recipients").
		WithArgs(99, 100).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	router := setupThresholdRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodDelete, "/threshold_recipients/99/100", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
