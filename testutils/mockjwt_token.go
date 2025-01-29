package testutils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// ✅ Known secret key for signing mock JWTs (Must match verification key)
const TestSecretKey = "test-secret"

// ✅ GenerateMockJWT creates a JWT token signed with HS256
func GenerateMockJWT() string {
	// 🔹 Create claims (payload)
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"name": "John Doe",
		"iat": time.Now().Unix(),    // Issued at
		"exp": time.Now().Add(time.Hour * 1).Unix(), // Expires in 1 hour
	}

	// 🔹 Create a token with HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 🔹 Sign the token using the known secret key
	tokenString, _ := token.SignedString([]byte(TestSecretKey))

	return tokenString
}
