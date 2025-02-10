package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

// CORSConfig applies CORS settings to allow the frontend to access the backend.
func CORSConfig(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"}, // ✅ Allow all origins for now (can be changed later to frontendURL)
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{"Content-Length", "Content-Type"},
			AllowCredentials: true,
		})
		return c.Handler(next)
	}
}
