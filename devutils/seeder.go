package devutils

import (
	"context"
	"log"
	"megga-backend/services"
)

// SeedDB seeds the database with test data.
func SeedDB() {
	seeds := map[string]string{
		"Inserting Users": `INSERT INTO users (email, first_name, last_name) VALUES 
			('testuser1@example.com', 'Test', 'User1'),
			('testuser2@example.com', 'Test', 'User2');`,

		"Inserting Recipients": `INSERT INTO recipients (email, first_name, last_name, designation) VALUES 
			('rep1@example.com', 'Jane', 'Doe', 'Representative'),
			('rep2@example.com', 'John', 'Smith', 'Governor');`,

		"Inserting Data": `INSERT INTO data (name, type, unit, previous_value, updated_value, last_updated, update_interval_in_days) VALUES 
			('Eggs', 'Good', 'USD per dozen', 3.25, 3.50, NOW(), 30),
			('Inflation', 'Indicator', '%', 2.5, 2.8, NOW(), 30);`,

		"Inserting Thresholds": `INSERT INTO thresholds (data_id, threshold_value, created_at) VALUES 
			(1, 5.0, NOW()),
			(2, 10.0, NOW());`,

		"Inserting Notifications": `INSERT INTO notifications (user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg) VALUES 
			(1, 1, 1, NOW(), 'Notification to User1', 'Notification to Recipient1'),
			(2, 2, 2, NOW(), 'Notification to User2', 'Notification to Recipient2');`,

		"Inserting Threshold Recipients": `INSERT INTO threshold_recipients (threshold_id, recipient_id, is_user) VALUES 
			(1, 1, true), -- User is notified
			(1, 1, false), -- Recipient1 is notified
			(2, 2, true), -- User is notified
			(2, 2, false); -- Recipient2 is notified`,
	}

	for description, query := range seeds {
		log.Printf("%s...", description)
		_, err := services.DB.Exec(context.Background(), query)
		if err != nil {
			log.Printf("Error seeding database: %v\n", err)
		}
	}
	log.Println("Database seeding completed successfully!")
}
