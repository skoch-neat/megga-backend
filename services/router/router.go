package router

import (
	"fmt"
	"net/http"
	"os"

	"megga-backend/handlers"
	"megga-backend/middleware"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the dynamic configuration for middleware
type Config struct {
	CognitoDomain string
	IDPURL        string
	TokenURL      string
	APIBaseURL    string
	FrontendURL   string
}

// loadConfig loads environment variables into the Config struct
func loadConfig() (*Config, error) {
	cfg := &Config{
		CognitoDomain: os.Getenv("COGNITO_DOMAIN"),
		IDPURL:        os.Getenv("COGNITO_IDP_URL"),
		TokenURL:      os.Getenv("COGNITO_TOKEN_URL"),
		APIBaseURL:    os.Getenv("API_BASE_URL"),
		FrontendURL:   os.Getenv("FRONTEND_URL"),
	}

	// Validate required fields
	if cfg.CognitoDomain == "" || cfg.IDPURL == "" || cfg.TokenURL == "" || cfg.APIBaseURL == "" || cfg.FrontendURL == "" {
		return nil, fmt.Errorf("missing one or more required environment variables")
	}

	return cfg, nil
}

// InitRouter initializes the router with API routes and middleware
func InitRouter(db *pgxpool.Pool) (http.Handler, error) {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Create the main router
	router := mux.NewRouter()

	// Handle OPTIONS preflight requests globally
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", cfg.FrontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK) // Respond with 200 OK for preflight
	})

	// Register API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handlers.GetUsers(w, r, db)
		} else if r.Method == "POST" {
			handlers.CreateUser(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "POST")

	// Apply middleware
	router.Use(middleware.LoggingMiddleware)           // Logs all requests globally
	router.Use(middleware.CORSConfig(cfg.FrontendURL)) // Handles CORS globally
	router.Use(middleware.CSPMiddleware(
		cfg.CognitoDomain, cfg.IDPURL, cfg.TokenURL, cfg.APIBaseURL, cfg.FrontendURL,
	)) // Handles CSP globally

	return router, nil
}
