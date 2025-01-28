package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pashagolub/pgxmock"
)

func TestRegisterUserRoutes(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := mux.NewRouter()
	RegisterUserRoutes(router, mock)

	tests := []struct {
		name          string
		method        string
		url           string
		expectedCode  int
		mockSetup     func()
		expectedBody  string
	}{
		{
			name:         "Valid GET /users",
			method:       "GET",
			url:          "/users",
			expectedCode: http.StatusOK,
			mockSetup: func() {
				mock.ExpectQuery("SELECT user_id, email, first_name, last_name FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "email", "first_name", "last_name"}).
						AddRow(1, "test@example.com", "Test", "User"))
			},
			expectedBody: `[{"user_id":1,"email":"test@example.com","first_name":"Test","last_name":"User"}]`,
		},
		{
			name:         "Valid POST /users",
			method:       "POST",
			url:          "/users",
			expectedCode: http.StatusCreated,
			mockSetup: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("test@example.com", "First", "Last").
					WillReturnRows(pgxmock.NewRows([]string{"user_id", "email"}).AddRow(1, "test@example.com"))
			},
			expectedBody: `{"message":"User created successfully","user":{"user_id":1,"email":"test@example.com","first_name":"First","last_name":"Last"}}`,
		},
		
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			req := httptest.NewRequest(tt.method, tt.url, nil)
			if tt.method == "POST" {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(`{"email":"test@example.com","first_name":"First","last_name":"Last"}`))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("[%s] Expected status %d, got %d", tt.name, tt.expectedCode, w.Code)
			}

			actualBody := strings.TrimSpace(w.Body.String())
			if actualBody != tt.expectedBody {
				t.Errorf("[%s] Expected body %q, got %q", tt.name, tt.expectedBody, actualBody)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("[%s] Unmet mock expectations: %v", tt.name, err)
			}
		})
	}
}
