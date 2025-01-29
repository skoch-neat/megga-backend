package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"megga-backend/testutils"

	"github.com/golang-jwt/jwt/v4"
)

// Define a unique type for context keys
type contextKey string

const claimsContextKey contextKey = "claims"

// CognitoConfig holds Cognito user pool details
type CognitoConfig struct {
	UserPoolID string
	Region     string
}

// ValidateCognitoToken is middleware to validate JWT tokens issued by AWS Cognito
func ValidateCognitoToken(config CognitoConfig) func(http.Handler) http.Handler {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", config.Region, config.UserPoolID)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			// ✅ Check if it's a test JWT using HS256
			if isMockJWT(tokenStr) {
				token, err := parseMockToken(tokenStr)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), claimsContextKey, token.Claims))
				next.ServeHTTP(w, r)
				return
			}

			// 🔹 Otherwise, process as a real Cognito JWT
			token, err := parseAndValidateToken(tokenStr, jwksURL)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), claimsContextKey, token.Claims))
			next.ServeHTTP(w, r)
		})
	}
}

// ✅ Check if the JWT is a mock test token
func isMockJWT(tokenStr string) bool {
	return strings.Contains(tokenStr, os.Getenv("MOCK_JWT_TOKEN"))
}

// ✅ Parse and Validate Mock JWT Token
func parseMockToken(tokenStr string) (*jwt.Token, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
	return parser.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(testutils.TestSecretKey), nil
	})
}

// parseAndValidateToken parses and validates a JWT token
func parseAndValidateToken(tokenStr, jwksURL string) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Validate token signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		// Fetch JWKS
		jwks, err := fetchJWKS(jwksURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		// Match key
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}
		key, ok := jwks[kid]
		if !ok {
			return nil, errors.New("key not found in JWKS")
		}

		return key, nil
	}

	return jwt.NewParser().Parse(tokenStr, keyFunc)
}

// fetchJWKS fetches and parses the JWKS and converts it to rsa.PublicKey
func fetchJWKS(jwksURL string) (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	// Map keys by kid
	keyMap := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		nBytes, err := decodeBase64(key.N)
		if err != nil {
			return nil, fmt.Errorf("failed to decode modulus: %w", err)
		}
		eBytes, err := decodeBase64ToInt(key.E)
		if err != nil {
			return nil, fmt.Errorf("failed to decode exponent: %w", err)
		}

		// Construct RSA Public Key
		pubKey := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: eBytes,
		}

		keyMap[key.Kid] = pubKey
	}

	return keyMap, nil
}

// decodeBase64 decodes a Base64 URL-encoded string to bytes
func decodeBase64(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}

// decodeBase64ToInt decodes a Base64 URL-encoded string to an integer
func decodeBase64ToInt(input string) (int, error) {
	decoded, err := decodeBase64(input)
	if err != nil {
		return 0, err
	}
	return int(new(big.Int).SetBytes(decoded).Uint64()), nil
}
