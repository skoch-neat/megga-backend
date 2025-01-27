package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
    dsn := os.Getenv("DATABASE_URI")
    var err error
    DB, err = pgxpool.New(context.Background(), dsn)
    if err != nil {
        log.Fatalf("Unable to connect to the database: %v", err)
    }
}
