package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"megga-backend/internal/config"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type CognitoConfig struct {
	UserPoolID string
	Region     string
}

func ValidateCognitoToken(cfg CognitoConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.IsDevelopmentMode() {
				log.Println("🔍 DEBUG: Entering ValidateCognitoToken middleware.")
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				if config.IsDevelopmentMode() {
					log.Println("❌ DEBUG: Missing Authorization header")
				}
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			if config.IsDevelopmentMode() {
				log.Println("🔍 DEBUG: Extracted token:", tokenStr)
			}

			token, err := parseAndValidateToken(tokenStr, cfg)
			if err != nil {
				if config.IsDevelopmentMode() {
					log.Println("❌ DEBUG: Token validation failed:", err)
				}
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				if config.IsDevelopmentMode() {
					log.Println("❌ DEBUG: Could not parse claims from token")
				}
				http.Error(w, "Unauthorized: Invalid claims", http.StatusUnauthorized)
				return
			}

			if config.IsDevelopmentMode() {
				log.Printf("🔍 DEBUG: Full extracted claims: %+v", claims)
			}

			email, emailOk := claims["email"].(string)
			firstName, firstNameOk := claims["given_name"].(string)
			lastName, lastNameOk := claims["family_name"].(string)

			if !emailOk || !firstNameOk || !lastNameOk {
				if config.IsDevelopmentMode() {
					log.Printf("❌ DEBUG: Missing claims in JWT. email: %v, firstName: %v, lastName: %v", email, firstName, lastName)
				}
				http.Error(w, "Missing required claims", http.StatusUnauthorized)
				return
			}

			r.Header.Set("X-User-Email", email)
			r.Header.Set("X-User-FirstName", firstName)
			r.Header.Set("X-User-LastName", lastName)

			if config.IsDevelopmentMode() {
				log.Printf("✅ DEBUG: JWT validated, claims injected. email=%s, firstName=%s, lastName=%s", email, firstName, lastName)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseAndValidateToken(tokenStr string, cfg CognitoConfig) (*jwt.Token, error) {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cfg.Region, cfg.UserPoolID)

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if config.IsDevelopmentMode() {
			log.Println("🔍 DEBUG: Validating JWT token.")
		}
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			if config.IsDevelopmentMode() {
				log.Println("❌ DEBUG: Unexpected signing method.")
			}
			return nil, errors.New("unexpected signing method")
		}

		jwks, err := fetchJWKS(jwksURL)
		if err != nil {
			if config.IsDevelopmentMode() {
				log.Printf("❌ DEBUG: Failed to fetch JWKS: %v", err)
			}
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			if config.IsDevelopmentMode() {
				log.Println("❌ DEBUG: Missing kid in token header.")
			}
			return nil, errors.New("missing kid in token header")
		}
		key, ok := jwks[kid]
		if !ok {
			if config.IsDevelopmentMode() {
				log.Println("❌ DEBUG: Key not found in JWKS.")
			}
			return nil, errors.New("key not found in JWKS")
		}
		if config.IsDevelopmentMode() {
			log.Println("✅ DEBUG: JWT validation successful.")
		}

		return key, nil
	}

	return jwt.Parse(tokenStr, keyFunc)
}

func fetchJWKS(jwksURL string) (map[string]*rsa.PublicKey, error) {
	if config.IsDevelopmentMode() {
		log.Printf("🔍 DEBUG: Fetching JWKS from URL: %s", jwksURL)
	}
	resp, err := http.Get(jwksURL)
	if err != nil {
		if config.IsDevelopmentMode() {
			log.Printf("❌ DEBUG: Failed to fetch JWKS: %v", err)
		}
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
		if config.IsDevelopmentMode() {
			log.Printf("❌ DEBUG: Failed to decode JWKS JSON: %v", err)
		}
		return nil, err
	}

	keyMap := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		nBytes, err := decodeBase64(key.N)
		if err != nil {
			if config.IsDevelopmentMode() {
				log.Printf("❌ DEBUG: Failed to decode modulus: %v", err)
			}
			return nil, fmt.Errorf("failed to decode modulus: %w", err)
		}
		eBytes, err := decodeBase64ToInt(key.E)
		if err != nil {
			if config.IsDevelopmentMode() {
				log.Printf("❌ DEBUG: Failed to decode exponent: %v", err)
			}
			return nil, fmt.Errorf("failed to decode exponent: %w", err)
		}

		pubKey := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: eBytes,
		}

		keyMap[key.Kid] = pubKey
	}

	if config.IsDevelopmentMode() {
		log.Println("✅ DEBUG: Successfully fetched and parsed JWKS.")
	}

	return keyMap, nil
}

func decodeBase64(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}

func decodeBase64ToInt(input string) (int, error) {
	decoded, err := decodeBase64(input)
	if err != nil {
		return 0, err
	}
	return int(new(big.Int).SetBytes(decoded).Uint64()), nil
}
