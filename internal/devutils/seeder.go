package devutils

import (
	"context"
	"log"
	"megga-backend/internal/database"
	"time"
)

func SeedDB(db database.DBQuerier) {
	seeds := []struct {
		description string
		query       string
	}{
		{"Inserting Users", `INSERT INTO users (email, first_name, last_name) VALUES 
			('user1@example.com', 'Alice', 'Smith'),
			('user2@example.com', 'Bob', 'Johnson')
			ON CONFLICT DO NOTHING;`},

		{"Inserting Recipients", `INSERT INTO recipients (email, first_name, last_name, designation) VALUES 
			('rep1@example.com', 'Jane', 'Doe', 'Representative'),
			('rep2@example.com', 'John', 'Smith', 'Governor')
			ON CONFLICT DO NOTHING;`},

		{"Inserting Data", `INSERT INTO data (name, series_id, unit, previous_value, latest_value, last_updated, period, year) VALUES 
			('Eggs', 'APU0000708111', 'per dozen', 3.25, 3.50, NOW(), 'M12', '2024'),
			('Inflation', 'LEU0252881600', 'constant 1982-1984 dollars', 2.5, 2.8, NOW(), 'Q4', '2024')
			ON CONFLICT DO NOTHING;`},

		{"Inserting Thresholds", `INSERT INTO thresholds (user_id, data_id, threshold_value, created_at, notify_user) VALUES 
			((SELECT user_id FROM users WHERE email = 'user1@example.com'), (SELECT data_id FROM data WHERE series_id = 'APU0000708111'), 5.0, NOW(), true),
			((SELECT user_id FROM users WHERE email = 'user2@example.com'), (SELECT data_id FROM data WHERE series_id = 'LEU0252881600'), 10.0, NOW(), false)
			ON CONFLICT DO NOTHING;`},

		{"Inserting Threshold Recipients", `INSERT INTO threshold_recipients (threshold_id, recipient_id) VALUES 
			((SELECT threshold_id FROM thresholds WHERE data_id = (SELECT data_id FROM data WHERE series_id = 'APU0000708111')), (SELECT recipient_id FROM recipients WHERE email = 'rep1@example.com')),
			((SELECT threshold_id FROM thresholds WHERE data_id = (SELECT data_id FROM data WHERE series_id = 'LEU0252881600')), (SELECT recipient_id FROM recipients WHERE email = 'rep2@example.com'))
			ON CONFLICT DO NOTHING;`},

		{"Inserting Notifications", `INSERT INTO notifications (user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg) VALUES 
			((SELECT user_id FROM users WHERE email = 'user1@example.com'), (SELECT recipient_id FROM recipients WHERE email = 'rep1@example.com'), (SELECT threshold_id FROM thresholds WHERE data_id = (SELECT data_id FROM data WHERE series_id = 'APU0000708111')), NOW(), 'Threshold Exceeded Alert', 'Threshold Notification for Recipient 1'),
			((SELECT user_id FROM users WHERE email = 'user2@example.com'), (SELECT recipient_id FROM recipients WHERE email = 'rep2@example.com'), (SELECT threshold_id FROM thresholds WHERE data_id = (SELECT data_id FROM data WHERE series_id = 'LEU0252881600')), NOW(), 'Threshold Exceeded Alert', 'Threshold Notification for Recipient 2')
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
