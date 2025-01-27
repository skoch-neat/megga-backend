package router

import (
	"fmt"
	"net/http"
	"os"

	"megga-backend/handlers"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
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

// cspMiddleware adds Content Security Policy headers dynamically
func cspMiddleware(cfg *Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csp := fmt.Sprintf(`
				default-src 'self';
				connect-src 'self' %s %s %s %s %s;
			`, cfg.FrontendURL, cfg.CognitoDomain, cfg.IDPURL, cfg.TokenURL, cfg.APIBaseURL)
			w.Header().Set("Content-Security-Policy", csp)
			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware sets up CORS configuration
func corsMiddleware(cfg *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return cors.New(cors.Options{
			AllowedOrigins:   []string{cfg.FrontendURL}, // Allow requests from the frontend
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
		}).Handler(next)
	}
}

// InitRouter initializes the router with API routes and middleware
func InitRouter(db *pgxpool.Pool) (http.Handler, error) {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()

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

	// Apply CSP middleware
	apiRouter.Use(cspMiddleware(cfg))

	// Apply CORS middleware
	return corsMiddleware(cfg)(router), nil
}
