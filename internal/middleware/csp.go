package middleware

import (
	"fmt"
	"net/http"
)

func CSPMiddleware(cognitoDomain, idpURL, tokenURL, apiBaseURL, frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csp := fmt.Sprintf(`
				default-src 'self';
				connect-src 'self' %s %s %s %s %s;
			`, frontendURL, cognitoDomain, idpURL, tokenURL, apiBaseURL)
			w.Header().Set("Content-Security-Policy", csp)
			next.ServeHTTP(w, r)
		})
	}
}
