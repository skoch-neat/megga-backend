package config

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

func init() {
	LoadEnv("env/.env.development")
}

func IsDevelopmentMode() bool {
	return os.Getenv("APP_ENV") == "development"
}

func IsProductionMode() bool {
	return os.Getenv("APP_ENV") == "production"
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
	env := os.Getenv("APP_ENV")
	envFile := ".env.development"

	if env == "production" {
		envFile = ".env.production"
	} else if env == "" {
		log.Println("⚠️ APP_ENV is not set, using default .env.development")
	}

	LoadEnv(envFile)

	ValidateEnv()
}

func LoadEnv(envFile string) {
	loadEnvFile(".env")
	loadEnvFile(envFile)
}

func loadEnvFile(envFile string) {
	err := godotenv.Load(envFile)
	if (IsDevelopmentMode()) {
		if err != nil {
			log.Printf("⚠️ No %s file found, using system environment variables", envFile)
		} else {
			log.Printf("✅ Loaded environment variables from %s", envFile)
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
