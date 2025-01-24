package services

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file and ensures required variables are set.
func LoadEnv() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Check if DATABASE_URI is set
	dsn := os.Getenv("DATABASE_URI")
	if dsn == "" {
		log.Fatal("DATABASE_URI is not set in the environment")
	}
}
