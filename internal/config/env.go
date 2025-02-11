package config

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

func init() {
	LoadEnv()
}

func IsDevelopmentMode() bool {
	return os.Getenv("APP_ENV") == "development"
}

func GetMockJWT() string {
	if !IsDevelopmentMode() {
		return ""
	}

	claims := jwt.MapClaims{
		"email":       "test@example.com",
		"given_name":  "TestFirst",
		"family_name": "TestLast",
		"exp":         time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	mockJWT, err := token.SignedString([]byte("mock-secret"))
	if err != nil {
		log.Fatalf("❌ Failed to generate mock JWT: %v", err)
	}

	return mockJWT
}

func LoadAndValidateEnv() {
	LoadEnv()
	ValidateEnv()
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if (IsDevelopmentMode()) {
		if err != nil {
			log.Printf("⚠️ No .env file found, using system environment variables")
		} else {
			log.Printf("✅ Loaded environment variables from .env")
		}
	}
}

func ValidateEnv() {
	requiredVars := []string{
		"DATABASE_URI",
		"COGNITO_CLIENT_ID",
		"BLS_API_KEY",
		"COGNITO_DOMAIN",
		"COGNITO_IDP_URL",
		"COGNITO_TOKEN_URL",
		"COGNITO_USER_POOL_ID",
		"AWS_REGION",
		"FRONTEND_URL",
		"API_BASE_URL",
		"PORT",
	}

	if IsDevelopmentMode() {
		requiredVars = append(requiredVars, "MOCK_JWT_TOKEN")
	}

	missingVars := []string{}

	for _, key := range requiredVars {
		if os.Getenv(key) == "" {
			log.Printf("Missing required environment variable: %s", key)
			missingVars = append(missingVars, key)
		}
	}

	if len(missingVars) > 0 {
		log.Fatalf("⛔ Startup failed due to missing environment variables: %v", missingVars)
	}

	log.Println("✅ All required environment variables are set!")
}
