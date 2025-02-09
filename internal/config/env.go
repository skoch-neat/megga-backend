package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func IsDevelopment() bool {
	return os.Getenv("ENVIRONMENT") == "development"
}

func LoadAndValidateEnv(envFile string) {
	LoadEnv(envFile)
	ValidateEnv()
}

func LoadEnv(envFile string) {
	loadEnvFile(".env")
	loadEnvFile(envFile)
}

func loadEnvFile(envFile string) {
	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("⚠️ No %s file found, using system environment variables", envFile)
	} else {
		log.Printf("✅ Loaded environment variables from %s", envFile)
	}
}

func ValidateEnv() {
	requiredVars := []string{
		"DATABASE_URI",
		"COGNITO_CLIENT_ID",
		"BLS_API_KEY",
		// "FRED_API_KEY",
		"COGNITO_DOMAIN",
		"COGNITO_IDP_URL",
		"COGNITO_TOKEN_URL",
		"COGNITO_USER_POOL_ID",
		"AWS_REGION",
		"FRONTEND_URL",
		"API_BASE_URL",
		"PORT",
	}

	if os.Getenv("ENVIRONMENT") == "development" {
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
