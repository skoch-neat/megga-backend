package devutils

import (
	"context"
	"log"
	"megga-backend/services/database"
	"time"
)

// SeedDB seeds the database with test data
func SeedDB(db database.DBQuerier) {
	seeds := map[string]string{
		"Inserting Users": `INSERT INTO users (email, first_name, last_name) VALUES 
			('testuser1@example.com', 'Test', 'User1'),
			('testuser2@example.com', 'Test', 'User2')
			ON CONFLICT DO NOTHING;`,

		"Inserting Recipients": `INSERT INTO recipients (email, first_name, last_name, designation) VALUES 
			('rep1@example.com', 'Jane', 'Doe', 'Representative'),
			('rep2@example.com', 'John', 'Smith', 'Governor')
			ON CONFLICT DO NOTHING;`,

		"Inserting Data": `INSERT INTO data (name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days) VALUES 
			('Eggs', 'APU0000708111', 'Good', 'USD per dozen', 3.25, 3.50, NOW(), 30),
			('Inflation', 'LEU0252881600', 'Indicator', '%', 2.5, 2.8, NOW(), 30)
			ON CONFLICT DO NOTHING;`,

		"Inserting Thresholds": `INSERT INTO thresholds (data_id, threshold_value, created_at) VALUES 
			(1, 5.0, NOW()),
			(2, 10.0, NOW())
			ON CONFLICT DO NOTHING;`,

		"Inserting Notifications": `INSERT INTO notifications (user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg) VALUES 
			(1, 1, 1, NOW(), 'Notification to User1', 'Notification to Recipient1'),
			(2, 2, 2, NOW(), 'Notification to User2', 'Notification to Recipient2')
			ON CONFLICT DO NOTHING;`,

		"Inserting Threshold Recipients": `INSERT INTO threshold_recipients (threshold_id, recipient_id, is_user) VALUES 
			(1, 1, true),
			(1, 1, false),
			(2, 2, true),
			(2, 2, false)
			ON CONFLICT DO NOTHING;`,
	}

	var failedSeeds []string

	for description, query := range seeds {
		log.Printf("[%s] %s...", time.Now().Format(time.RFC3339), description)
		_, err := db.Exec(context.Background(), query)
		if err != nil {
			log.Printf("Error seeding database (%s): %v", description, err)
			failedSeeds = append(failedSeeds, description)
		}
	}

	if len(failedSeeds) > 0 {
		log.Printf("Seeding completed with errors for: %v", failedSeeds)
	} else {
		log.Println("Database seeding completed successfully!")
	}
}
