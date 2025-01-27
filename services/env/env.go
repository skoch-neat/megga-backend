package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

// ValidateEnv ensures all required environment variables are set.
func ValidateEnv() {
	requiredVars := []string{
		"DATABASE_URI",
		"PORT",
		"BLS_API_KEY",
		"FRED_API_KEY",
		"COGNITO_CLIENT_ID",
		"COGNITO_DOMAIN",
		"COGNITO_IDP_URL",
		"COGNITO_TOKEN_URL",
		"API_BASE_URL",
		"FRONTEND_URL",
	}

	for _, key := range requiredVars {
		if os.Getenv(key) == "" {
			log.Fatalf("Environment variable %s is not set", key)
		}
	}
	log.Println("All required environment variables are set!")
}
