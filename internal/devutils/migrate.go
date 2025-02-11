package devutils

import (
	"context"
	"log"
	"megga-backend/internal/database"
)

func MigrateDB(db database.DBQuerier) {
	migrations := []struct {
		description string
		sql         string
	}{
		{"Creating User table", `CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255)
		)`},
		{"Creating Recipient table", `CREATE TABLE IF NOT EXISTS recipients (
			recipient_id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			designation VARCHAR(255)
		)`},
		{"Creating Data table", `CREATE TABLE IF NOT EXISTS data (
			data_id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			series_id VARCHAR(255) UNIQUE NOT NULL,
			unit VARCHAR(50),
			previous_value FLOAT,
			latest_value FLOAT,
			last_updated TIMESTAMP DEFAULT NOW(),
			period VARCHAR(10),
			year VARCHAR(10)
		)`},
		{"Creating Threshold table", `CREATE TABLE IF NOT EXISTS thresholds (
			threshold_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			data_id INT NOT NULL REFERENCES data(data_id) ON DELETE CASCADE,
			threshold_value FLOAT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			notify_user BOOLEAN DEFAULT FALSE
		)`},
		{"Creating Threshold_Recipient table", `CREATE TABLE IF NOT EXISTS threshold_recipients (
			threshold_id INT NOT NULL REFERENCES thresholds(threshold_id) ON DELETE CASCADE,
			recipient_id INT NOT NULL REFERENCES recipients(recipient_id) ON DELETE CASCADE,
			PRIMARY KEY (threshold_id, recipient_id)
		)`},
		{"Creating Notification table", `CREATE TABLE IF NOT EXISTS notifications (
			notification_id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			recipient_id INT NOT NULL REFERENCES recipients(recipient_id) ON DELETE CASCADE,
			threshold_id INT NOT NULL REFERENCES thresholds(threshold_id) ON DELETE CASCADE,
			sent_at TIMESTAMP,
			user_msg TEXT,
			recipient_msg TEXT
		)`},
	}

	for _, m := range migrations {
		log.Printf("%s...", m.description)
		_, err := db.Exec(context.Background(), m.sql)
		if err != nil {
			log.Fatalf("Failed to run migration (%s): %v", m.description, err)
		}
	}
	log.Println("Database migration completed successfully.")
}
