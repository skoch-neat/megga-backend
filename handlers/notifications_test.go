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

// Helper function to set up the router with proper route extraction
func setupNotificationRouter(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	RegisterNotificationRoutes(router, mock)
	return router
}

// TestCreateNotification - Ensures a notification can be created
func TestCreateNotification(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("INSERT INTO notifications").
		WithArgs(1, 2, 3, "User message", "Recipient message").
		WillReturnRows(pgxmock.NewRows([]string{"notification_id"}).AddRow(42))

	router := setupNotificationRouter(mock)

	body := bytes.NewBufferString(`{
		"user_id": 1,
		"recipient_id": 2,
		"threshold_id": 3,
		"user_msg": "User message",
		"recipient_msg": "Recipient message"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/notifications", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

// TestGetNotifications - Ensures all notifications can be retrieved
func TestGetNotifications(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT notification_id, user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg FROM notifications").
		WillReturnRows(pgxmock.NewRows([]string{"notification_id", "user_id", "recipient_id", "threshold_id", "sent_at", "user_msg", "recipient_msg"}).
			AddRow(42, 1, 2, 3, time.Now(), "User message", "Recipient message"))

	router := setupNotificationRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestGetNotificationByID_Success - Ensures a notification can be retrieved by ID
func TestGetNotificationByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery("SELECT notification_id, user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg FROM notifications WHERE notification_id =").
		WithArgs(42).
		WillReturnRows(pgxmock.NewRows([]string{"notification_id", "user_id", "recipient_id", "threshold_id", "sent_at", "user_msg", "recipient_msg"}).
			AddRow(42, 1, 2, 3, time.Now(), "User message", "Recipient message"))

	router := setupNotificationRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/notifications/42", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestGetNotificationByID_NotFound - Ensures 404 response for non-existent notification
func TestGetNotificationByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(`
	SELECT notification_id, user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg
	FROM notifications WHERE notification_id = \$1`).
	WithArgs(99).
	WillReturnError(pgx.ErrNoRows)


	router := setupNotificationRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/notifications/99", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestUpdateNotification - Ensures a notification can be updated
func TestUpdateNotification(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("UPDATE notifications").
		WithArgs("Updated user message", "Updated recipient message", 42).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	router := setupNotificationRouter(mock)

	body := bytes.NewBufferString(`{
		"user_msg": "Updated user message",
		"recipient_msg": "Updated recipient message"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/notifications/42", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestDeleteNotification_Success - Ensures a notification can be deleted
func TestDeleteNotification_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM notifications").
		WithArgs(42).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	router := setupNotificationRouter(mock)

	req := httptest.NewRequest(http.MethodDelete, "/notifications/42", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestDeleteNotification_NotFound - Ensures 404 response for deleting a non-existent notification
func TestDeleteNotification_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectExec("DELETE FROM notifications").
		WithArgs(99). // Non-existent ID
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	router := setupNotificationRouter(mock)

	req := httptest.NewRequest(http.MethodDelete, "/notifications/99", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
