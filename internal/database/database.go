package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	dsn := os.Getenv("DATABASE_URI")
	dbPool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = dbPool
	log.Println("🔗 Database connection established!")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("🚪 Database connection closed")
	}
}
