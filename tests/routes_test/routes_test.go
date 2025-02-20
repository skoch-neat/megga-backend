package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pashagolub/pgxmock"
	"megga-backend/handlers"
)

func setupRouter(mock pgxmock.PgxPoolIface) *mux.Router {
	router := mux.NewRouter()
	handlers.RegisterDataRoutes(router, mock)
	handlers.RegisterThresholdRoutes(router, mock)
	return router
}

func TestRegisterDataRoutes(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	router := setupRouter(mock)

	validRoutes := []struct {
		method string
		url    string
	}{
		{"POST", "/data"},
		{"GET", "/data"},
		{"GET", "/data/1"},
		{"PUT", "/data/1"},
		{"DELETE", "/data/1"},
	}
	for _, route := range validRoutes {
		t.Run(route.method+" "+route.url, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusMethodNotAllowed {
				t.Errorf("Expected %s %s to be allowed, got 405", route.method, route.url)
			}
		})
	}

	req := httptest.NewRequest("PATCH", "/data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected PATCH /data to be 405, got %d", w.Code)
	}
}