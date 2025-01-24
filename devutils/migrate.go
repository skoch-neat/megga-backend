package devutils

import (
	"context"
	"log"
	"megga-backend/services"
)

// MigrateDB runs database schema migrations.
func MigrateDB() {
	// Use the global database connection pool from services
	db := services.DB

	migrations := map[string]string{
		"Creating User table": `CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255)
		)`,
		"Creating Recipient table": `CREATE TABLE IF NOT EXISTS recipients (
			recipient_id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			designation VARCHAR(255)
		)`,
		"Creating Threshold table": `CREATE TABLE IF NOT EXISTS thresholds (
			threshold_id SERIAL PRIMARY KEY,
			data_id INT NOT NULL,
			threshold_value FLOAT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)`,
		"Creating Data table": `CREATE TABLE IF NOT EXISTS data (
			data_id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(255) NOT NULL,
			unit VARCHAR(50),
			previous_value FLOAT,
			updated_value FLOAT,
			last_updated TIMESTAMP,
			update_interval_in_days INT
		)`,
		"Creating Notification table": `CREATE TABLE IF NOT EXISTS notifications (
			notification_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			recipient_id INT NOT NULL,
			threshold_id INT NOT NULL,
			sent_at TIMESTAMP NOT NULL,
			user_msg TEXT,
			recipient_msg TEXT
		)`,
		"Creating Threshold_Recipient table": `CREATE TABLE IF NOT EXISTS threshold_recipients (
			threshold_id INT NOT NULL,
			recipient_id INT NOT NULL,
			is_user BOOLEAN NOT NULL,
			PRIMARY KEY (threshold_id, recipient_id, is_user)
		)`,
	}

	for description, sql := range migrations {
		log.Printf("%s...", description)
		_, err := db.Exec(context.Background(), sql)
		if err != nil {
			log.Fatalf("Failed to run migration (%s): %v", description, err)
		}
	}
	log.Println("Database migration completed successfully.")
}
