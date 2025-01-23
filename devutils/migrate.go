package devutils

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func connectToDB() *pgxpool.Pool {
	dsn := "postgres://<username>:<password>@localhost:5432/megga_dev?sslmode=disable"
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func MigrateDB() {
	DB = connectToDB()
	defer DB.Close()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS recipients (
			recipient_id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			designation VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS thresholds (
			threshold_id SERIAL PRIMARY KEY,
			data_id INT NOT NULL,
			threshold_value FLOAT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS data (
			data_id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(255) NOT NULL,
			unit VARCHAR(50),
			previous_value FLOAT,
			updated_value FLOAT,
			last_updated TIMESTAMP,
			update_interval_in_days INT
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			notification_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			recipient_id INT NOT NULL,
			threshold_id INT NOT NULL,
			sent_at TIMESTAMP NOT NULL,
			user_msg TEXT,
			recipient_msg TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS threshold_recipients (
			threshold_id INT NOT NULL,
			recipient_id INT NOT NULL,
			is_user BOOLEAN NOT NULL,
			PRIMARY KEY (threshold_id, recipient_id, is_user)
		)`,
	}

	for _, sql := range migrations {
		_, err := DB.Exec(context.Background(), sql)
		if err != nil {
			log.Fatalf("Failed to run migration: %v", err)
		}
	}
	log.Println("Database migration completed successfully.")
}
