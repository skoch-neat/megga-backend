package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	// "github.com/joho/godotenv"
)

var DB *pgxpool.Pool

func InitDB() error {
	dsn := os.Getenv("DATABASE_URI")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URI is not set")
		return nil
	}

	dbPool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return err
	}

	DB = dbPool
	log.Println("🔗 Database connection established!")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		DB = nil
		log.Println("🚪 Database connection closed")
	}
}
