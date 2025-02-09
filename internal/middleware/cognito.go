package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"megga-backend/internal/config"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

type CognitoConfig struct {
	UserPoolID string
	Region     string
}

func ValidateCognitoToken(cfg CognitoConfig) func(http.Handler) http.Handler {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cfg.Region, cfg.UserPoolID)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("❌ DEBUG: Missing Authorization header")
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			log.Println("🔍 DEBUG: Extracted token:", tokenStr)

			// Mock JWT in dev mode
			if config.IsDevelopmentMode() {
				mockToken := config.GetMockJWT()
				if mockToken != "" && tokenStr == mockToken {
					log.Println("⚠️ DEBUG: Using MOCK_JWT_TOKEN in development mode")
					mockClaims, err := generateMockClaims()
					if err != nil {
						log.Println("❌ DEBUG: Invalid mock token:", err)
						http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
						return
					}

					// Inject claims into headers instead of context
					r.Header.Set("X-User-Email", mockClaims["email"].(string))
					r.Header.Set("X-User-FirstName", mockClaims["given_name"].(string))
					r.Header.Set("X-User-LastName", mockClaims["family_name"].(string))

					next.ServeHTTP(w, r)
					return
				}
			}

			// Validate real JWT
			token, err := parseAndValidateToken(tokenStr, jwksURL)
			if err != nil {
				log.Println("❌ DEBUG: Token validation failed:", err)
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Println("❌ DEBUG: Could not parse claims from token")
				http.Error(w, "Unauthorized: Invalid claims", http.StatusUnauthorized)
				return
			}

			// Inject claims into headers
			r.Header.Set("X-User-Email", claims["email"].(string))
			r.Header.Set("X-User-FirstName", claims["given_name"].(string))
			r.Header.Set("X-User-LastName", claims["family_name"].(string))

			log.Println("✅ DEBUG: JWT validated, claims injected into headers.")
			next.ServeHTTP(w, r)
		})
	}
}

func parseAndValidateToken(tokenStr, jwksURL string) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		jwks, err := fetchJWKS(jwksURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

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

		pubKey := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: eBytes,
		}

		keyMap[key.Kid] = pubKey
	}

	return keyMap, nil
}

// ✅ Parse the mock JWT manually and return a token with claims
func generateMockClaims() (jwt.MapClaims, error) {
	claims := jwt.MapClaims{
		"email":       "test@example.com",
		"given_name":  "TestFirst",
		"family_name": "TestLast",
	}

	// 🔹 Create and sign the mock JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("mock-secret")) // Use a dummy secret
	if err != nil {
		return nil, fmt.Errorf("failed to sign mock token: %w", err)
	}

	// 🔹 Parse the signed token
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("mock-secret"), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse mock token: %w", err)
	}

	// 🔹 Extract and return claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		return claims, nil
	} else {
		return nil, errors.New("invalid mock token claims")
	}
}

// ✅ Decode base64 string
func decodeBase64(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}

// ✅ Decode base64 string to integer
func decodeBase64ToInt(input string) (int, error) {
	decoded, err := decodeBase64(input)
	if err != nil {
		return 0, err
	}
	return int(new(big.Int).SetBytes(decoded).Uint64()), nil
}
