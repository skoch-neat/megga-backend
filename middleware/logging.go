package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each incoming HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log the request details
		log.Printf(
			"%s %s %s %d %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			lrw.statusCode,
			time.Since(start),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
