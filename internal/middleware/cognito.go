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

	"github.com/golang-jwt/jwt/v4"
)

type CognitoConfig struct {
	UserPoolID string
	Region     string
}

func ValidateCognitoToken(cfg CognitoConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Println("🔍 DEBUG: Entering ValidateCognitoToken middleware.")

            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                log.Println("❌ DEBUG: Missing Authorization header")
                http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
                return
            }

            tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
            log.Println("🔍 DEBUG: Extracted token:", tokenStr)

            // Validate JWT token
            token, err := parseAndValidateToken(tokenStr, cfg)
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

            // Log full claims for debugging
            log.Printf("🔍 DEBUG: Full extracted claims: %+v", claims)

            // Extract claims safely
            email, emailOk := claims["email"].(string)
            firstName, firstNameOk := claims["given_name"].(string)
            lastName, lastNameOk := claims["family_name"].(string)

            // If "email" claim is missing, try "username" instead (Cognito sometimes uses this)
            if !emailOk {
                email, emailOk = claims["username"].(string)
            }

            // Ensure required claims exist
            if !emailOk || !firstNameOk || !lastNameOk {
                log.Printf("❌ DEBUG: Missing claims in JWT. email: %v, firstName: %v, lastName: %v", email, firstName, lastName)
                http.Error(w, "Missing required claims", http.StatusUnauthorized)
                return
            }

            // Inject claims into headers
            r.Header.Set("X-User-Email", email)
            r.Header.Set("X-User-FirstName", firstName)
            r.Header.Set("X-User-LastName", lastName)

            log.Printf("✅ DEBUG: JWT validated, claims injected. email=%s, firstName=%s, lastName=%s", email, firstName, lastName)
            next.ServeHTTP(w, r)
        })
    }
}

func parseAndValidateToken(tokenStr string, cfg CognitoConfig) (*jwt.Token, error) {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cfg.Region, cfg.UserPoolID)

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

	return jwt.Parse(tokenStr, keyFunc)
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
