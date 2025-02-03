package middleware

import (
	"github.com/rs/cors"
	"net/http"
	"log"
)

func CORSConfig(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log.Printf("CORS middleware applied for origin: %s", frontendURL)
		return cors.New(cors.Options{
			AllowedOrigins:   []string{frontendURL},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{"Content-Length", "Content-Type"},
			AllowCredentials: true,
		}).Handler(next)
	}
}
