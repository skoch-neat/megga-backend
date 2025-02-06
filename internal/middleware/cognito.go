package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const claimsContextKey contextKey = "claims"

type CognitoConfig struct {
	UserPoolID string
	Region     string
}

func ValidateCognitoToken(config CognitoConfig) func(http.Handler) http.Handler {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", config.Region, config.UserPoolID)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("❌ Missing Authorization header")
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := parseAndValidateToken(tokenStr, jwksURL)
			if err != nil {
				log.Println("❌ Token validation failed:", err)
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), claimsContextKey, token.Claims))
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
