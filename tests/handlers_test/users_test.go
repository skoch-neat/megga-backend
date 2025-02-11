package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"

	"megga-backend/handlers"
	"megga-backend/internal/config"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
)

var MOCK_JWT_TOKEN = config.GetMockJWT()

func setupRouterWithoutMiddleware(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	handlers.RegisterUserRoutes(router, mock)
	return router
}

func setupRouterWithMiddleware(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("🛠 Middleware injecting headers: email=%s, firstName=%s, lastName=%s",
				r.Header.Get("X-User-Email"),
				r.Header.Get("X-User-FirstName"),
				r.Header.Get("X-User-LastName"),
			)
			next.ServeHTTP(w, r)
		})
	})

	handlers.RegisterUserRoutes(router, mock)
	return router
}

func TestCreateUser_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := setupRouterWithMiddleware(mock)

	mock.ExpectQuery(`SELECT user_id FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectQuery(`INSERT INTO users \(email, first_name, last_name\) VALUES \(\$1, \$2, \$3\) RETURNING user_id, email, first_name, last_name`).
		WithArgs("test@example.com", "First", "Last").
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "email", "first_name", "last_name"}).
			AddRow(1, "test@example.com", "First", "Last"))

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(`{
		"email": "test@example.com",
		"first_name": "First",
		"last_name": "Last"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+MOCK_JWT_TOKEN)

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

	router := setupRouterWithMiddleware(mock)

	mock.ExpectQuery(`SELECT user_id FROM users WHERE email = \$1`).
		WithArgs("existing@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(2))

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(`{
		"email": "existing@example.com",
		"first_name": "Existing",
		"last_name": "User"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+MOCK_JWT_TOKEN)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var expected map[string]interface{}
	var actual map[string]interface{}

	expectedJSON := `{"message":"User already exists","user":{"user_id":2,"email":"existing@example.com","first_name":"Existing","last_name":"User"}}`
	json.Unmarshal([]byte(expectedJSON), &expected)
	json.Unmarshal(w.Body.Bytes(), &actual)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected response:\n%v\nGot:\n%v", expected, actual)
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

func TestGetUserByEmail_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, email, first_name, last_name FROM users WHERE LOWER(email) = LOWER($1)`)).
		WithArgs("test@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "email", "first_name", "last_name"}).
			AddRow(1, "test@example.com", "John", "Doe"))

	req := httptest.NewRequest("GET", "/users/test@example.com", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+MOCK_JWT_TOKEN)
	req.Header.Set("X-User-Email", "test@example.com")
	req.Header.Set("X-User-FirstName", "TestFirst")
	req.Header.Set("X-User-LastName", "TestLast")

	w := httptest.NewRecorder()
	router := setupRouterWithMiddleware(mock)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, email, first_name, last_name FROM users WHERE LOWER(email) = LOWER($1)`)).
		WithArgs("notfound@example.com").
		WillReturnError(pgx.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING user_id, email, first_name, last_name`)).
		WithArgs("notfound@example.com", "TestFirstName", "TestLastName").
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "email", "first_name", "last_name"}).
			AddRow(3, "notfound@example.com", "TestFirstName", "TestLastName"))

	req := httptest.NewRequest("GET", "/users/notfound@example.com", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+MOCK_JWT_TOKEN)
	req.Header.Set("X-User-Email", "notfound@example.com")
	req.Header.Set("X-User-FirstName", "TestFirstName")
	req.Header.Set("X-User-LastName", "TestLastName")

	w := httptest.NewRecorder()
	router := setupRouterWithMiddleware(mock)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201 (Created), got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestGetThresholdsForUser_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
    SELECT t.threshold_id, t.data_id, d.name, t.threshold_value, t.notify_user,
       	COALESCE(ARRAY_AGG(tr.recipient_id) FILTER (WHERE tr.recipient_id IS NOT NULL), ARRAY[]::BIGINT[]) AS recipients
	FROM thresholds t
	JOIN data d ON t.data_id = d.data_id
	LEFT JOIN threshold_recipients tr ON t.threshold_id = tr.threshold_id
	WHERE t.user_id = $1
	GROUP BY t.threshold_id, d.name`)).
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "data_id", "name", "threshold_value", "notify_user", "recipients"}).
			AddRow(101, 1, "Eggs", 10.0, true, []int64{1, 2}).
			AddRow(102, 2, "Milk", 15.5, false, []int64{3}))

	req := httptest.NewRequest("GET", "/users/1/thresholds", nil)
	w := httptest.NewRecorder()
	router := setupRouterWithoutMiddleware(mock)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetThresholdsForUser_NoThresholds(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
    SELECT t.threshold_id, t.data_id, d.name, t.threshold_value, t.notify_user,
       	COALESCE(ARRAY_AGG(tr.recipient_id) FILTER (WHERE tr.recipient_id IS NOT NULL), ARRAY[]::BIGINT[]) AS recipients
	FROM thresholds t
	JOIN data d ON t.data_id = d.data_id
	LEFT JOIN threshold_recipients tr ON t.threshold_id = tr.threshold_id
	WHERE t.user_id = $1
	GROUP BY t.threshold_id, d.name`)).
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"threshold_id", "data_id", "threshold_value"}))

	req := httptest.NewRequest("GET", "/users/1/thresholds", nil)
	w := httptest.NewRecorder()
	router := setupRouterWithoutMiddleware(mock)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteAllThresholdsForUser_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM threshold_recipients WHERE threshold_id IN (SELECT threshold_id FROM thresholds WHERE user_id = $1)`)).
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("DELETE", 2))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM thresholds WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("DELETE", 2))

	mock.ExpectCommit()

	req := httptest.NewRequest("DELETE", "/users/1/thresholds", nil)
	w := httptest.NewRecorder()
	router := setupRouterWithoutMiddleware(mock)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestDeleteAllThresholdsForUser_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM threshold_recipients WHERE threshold_id IN (SELECT threshold_id FROM thresholds WHERE user_id = $1)`)).
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("DELETE", 2))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM thresholds WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	mock.ExpectRollback()

	req := httptest.NewRequest("DELETE", "/users/1/thresholds", nil)
	w := httptest.NewRecorder()
	router := setupRouterWithoutMiddleware(mock)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRegisterUserRoutes(t *testing.T) {
	router := mux.NewRouter()
	mock, _ := pgxmock.NewPool()
	handlers.RegisterUserRoutes(router, mock)
}
