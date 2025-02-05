package devutils

import (
	"context"
	"log"
	"megga-backend/services/database"
	"time"
)

func SeedDB(db database.DBQuerier) {
	seeds := []struct {
		description string
		query       string
	}{
		{"Inserting Users", `INSERT INTO users (email, first_name, last_name) VALUES 
			('testuser1@example.com', 'Test', 'User1'),
			('testuser2@example.com', 'Test', 'User2'),
			('sarahkoch810@gmail.com', 'Sarah', 'Koch')
			ON CONFLICT DO NOTHING;`}, // Delete my info after testing

		{"Inserting Recipients", `INSERT INTO recipients (email, first_name, last_name, designation) VALUES 
			('rep1@example.com', 'Jane', 'Doe', 'Representative'),
			('rep2@example.com', 'John', 'Smith', 'Governor')
			ON CONFLICT DO NOTHING;`},

		{"Inserting Data", `INSERT INTO data (name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days) VALUES 
			('Eggs', 'APU0000708111', 'Good', 'USD per dozen', 3.25, 3.50, NOW(), 30),
			('Inflation', 'LEU0252881600', 'Indicator', '%', 2.5, 2.8, NOW(), 30)
			ON CONFLICT DO NOTHING;`},

		{"Inserting Thresholds", `INSERT INTO thresholds (user_id, data_id, threshold_value, created_at, notify_user) VALUES 
			(1, 1, 5.0, NOW(), true),
			(2, 2, 10.0, NOW(), false)
			ON CONFLICT DO NOTHING;`},

		{"Inserting Threshold Recipients", `INSERT INTO threshold_recipients (threshold_id, recipient_id) VALUES 
			(1, 1),
			(1, 2),
			(2, 2),
			(2, 1)
			ON CONFLICT DO NOTHING;`},

		{"Inserting Notifications", `INSERT INTO notifications (user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg) VALUES 
			(1, 1, 1, NOW(), 'Notification to User1', 'Notification to Recipient1'),
			(2, 2, 2, NOW(), 'Notification to User2', 'Notification to Recipient2')
			ON CONFLICT DO NOTHING;`},
	}

	var failedSeeds []string

	for _, seed := range seeds {
		log.Printf("[%s] %s...", time.Now().Format(time.RFC3339), seed.description)
		_, err := db.Exec(context.Background(), seed.query)
		if err != nil {
			log.Printf("Error seeding database (%s): %v", seed.description, err)
			failedSeeds = append(failedSeeds, seed.description)
		}
	}

	if len(failedSeeds) > 0 {
		log.Printf("Seeding completed with errors for: %v", failedSeeds)
	} else {
		log.Println("Database seeding completed successfully!")
	}
}
