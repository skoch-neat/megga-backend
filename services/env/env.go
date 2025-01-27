package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file and ensures required variables are set.
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

// ValidateEnv ensures all required environment variables are set.
func ValidateEnv() {
	requiredVars := []string{"DATABASE_URI", "PORT", "BLS_API_KEY", "FRED_API_KEY", "COGNITO_CLIENT_ID"}
	for _, key := range requiredVars {
		if os.Getenv(key) == "" {
			log.Fatalf("Environment variable %s is not set", key)
		}
	}
	log.Println("All required environment variables are set!")
}
