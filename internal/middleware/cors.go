package middleware

import (
	"log"
	"megga-backend/internal/config"
	"net/http"

	"github.com/rs/cors"
)

func CORSConfig(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsDevelopmentMode() {
				log.Printf("🔍 DEBUG: Request Method = %s", r.Method)
				log.Printf("🔍 DEBUG: Request Headers = %+v", r.Header)
				log.Printf("🔍 DEBUG: MethodOptions = %s", http.MethodOptions)
			}

			// 🔴 Set CORS headers for ALL responses
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// ✅ Explicitly handle preflight (OPTIONS) requests
			if r.Method == http.MethodOptions {
				if config.IsDevelopmentMode() {
					log.Println("✅ DEBUG: Handling CORS preflight request (OPTIONS)")
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Apply normal CORS middleware for other requests
			cors.New(cors.Options{
				AllowedOrigins:   []string{frontendURL},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				ExposedHeaders:   []string{"Content-Length", "Content-Type"},
				AllowCredentials: true,
			}).Handler(next).ServeHTTP(w, r)
		})
	}
}
