package handlers_test

import (
	"bytes"
	"megga-backend/handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pashagolub/pgxmock"
)

func setupRecipientRouter(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	handlers.RegisterRecipientRoutes(router, mock)
	return router
}

func TestCreateRecipient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("INSERT INTO recipients").
		WithArgs("test@example.com", "John", "Doe", "Representative").
		WillReturnRows(pgxmock.NewRows([]string{"recipient_id"}).AddRow(42))

	router := setupRecipientRouter(mock)

	body := bytes.NewBufferString(`{
		"email": "test@example.com",
		"first_name": "John",
		"last_name": "Doe",
		"designation": "Representative"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/recipients", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestGetRecipients(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT recipient_id, email, first_name, last_name, designation FROM recipients").
		WillReturnRows(pgxmock.NewRows([]string{"recipient_id", "email", "first_name", "last_name", "designation"}).
			AddRow(42, "test@example.com", "John", "Doe", "Representative"))

	router := setupRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/recipients", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetRecipientByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT recipient_id, email, first_name, last_name, designation FROM recipients WHERE recipient_id =").
		WithArgs(42).
		WillReturnRows(pgxmock.NewRows([]string{"recipient_id", "email", "first_name", "last_name", "designation"}).
			AddRow(42, "test@example.com", "John", "Doe", "Representative"))

	router := setupRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/recipients/42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateRecipient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("UPDATE recipients").
		WithArgs("updated@example.com", "Jane", "Smith", "Updated Role", 42).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	router := setupRecipientRouter(mock)

	body := bytes.NewBufferString(`{
		"email": "updated@example.com",
		"first_name": "Jane",
		"last_name": "Smith",
		"designation": "Updated Role"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/recipients/42", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteRecipient_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM recipients").
		WithArgs(42).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	router := setupRecipientRouter(mock)

	req := httptest.NewRequest(http.MethodDelete, "/recipients/42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
