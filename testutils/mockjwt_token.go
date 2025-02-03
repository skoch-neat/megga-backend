package testutils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TestSecretKey = "test-secret"

func GenerateMockJWT() string {
	claims := jwt.MapClaims{
		"sub": "1234567890",
		"name": "John Doe",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, _ := token.SignedString([]byte(TestSecretKey))

	return tokenString
}
