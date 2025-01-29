package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"megga-backend/middleware"
	"megga-backend/testutils"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

var mockJWT = "Bearer " + testutils.GenerateMockJWT()

func setupRouterWithoutMiddleware(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	RegisterUserRoutes(router, mock)
	return router
}

func setupRouterWithMiddleware(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.ValidateCognitoToken(middleware.CognitoConfig{
		UserPoolID: "test-pool-id",
		Region:     "us-east-1",
	}))
	RegisterUserRoutes(router, mock)
	return router
}

func TestGetUsers(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := setupRouterWithoutMiddleware(mock)

	mock.ExpectQuery(`SELECT user_id, email, first_name, last_name FROM users`).
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "email", "first_name", "last_name"}).AddRow(1, "test@example.com", "Test", "User"))

	req := httptest.NewRequest("GET", "/users", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCreateUser_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := setupRouterWithoutMiddleware(mock)

	mock.ExpectQuery(`SELECT user_id FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectQuery(`INSERT INTO users \(email, first_name, last_name\) VALUES \(\$1, \$2, \$3\) RETURNING user_id, email`).
		WithArgs("test@example.com", "First", "Last").
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "email"}).AddRow(1, "test@example.com"))

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(`{
		"email": "test@example.com",
		"first_name": "First",
		"last_name": "Last"
	}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestCreateUser_UserAlreadyExists(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := setupRouterWithoutMiddleware(mock)

	mock.ExpectQuery(`SELECT user_id FROM users WHERE email = \$1`).
		WithArgs("existing@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(2))

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(`{
		"email": "existing@example.com",
		"first_name": "Existing",
		"last_name": "User"
	}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if strings.TrimSpace(w.Body.String()) != `{"message":"User already exists","user":{"user_id":2,"email":"existing@example.com","first_name":"Existing","last_name":"User"}}` {
		t.Errorf("Expected correct response, got %q", w.Body.String())
	}
}

func TestCreateUser_InvalidData(t *testing.T) {
	router := setupRouterWithoutMiddleware(nil)

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(`{
		"email": "invalid@example.com"
	}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUserHandlers_WithMiddleware(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := setupRouterWithMiddleware(mock)

	mock.ExpectQuery(`SELECT user_id FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectQuery(`INSERT INTO users \(email, first_name, last_name\) VALUES \(\$1, \$2, \$3\) RETURNING user_id, email`).
		WithArgs("test@example.com", "First", "Last").
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "email"}).AddRow(1, "test@example.com"))

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(`{
		"email": "test@example.com",
		"first_name": "First",
		"last_name": "Last"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mockJWT)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestRegisterUserRoutes(t *testing.T) {
	router := mux.NewRouter()
	mock, _ := pgxmock.NewPool()
	RegisterUserRoutes(router, mock)
}