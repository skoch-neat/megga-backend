package middleware

import (
	"github.com/rs/cors"
	"net/http"
	"log"
)

// CORSConfig sets up CORS middleware with the specified frontend URL
func CORSConfig(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log.Printf("CORS middleware applied for origin: %s", frontendURL)
		return cors.New(cors.Options{
			AllowedOrigins:   []string{frontendURL}, // Allow requests from the specified frontend
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Include OPTIONS for preflight
			AllowedHeaders:   []string{"Content-Type", "Authorization"}, // Accept content type and authorization headers
			ExposedHeaders:   []string{"Content-Length", "Content-Type"}, // Optionally expose additional headers
			AllowCredentials: true, // Support cookies or authentication credentials
		}).Handler(next)
	}
}
