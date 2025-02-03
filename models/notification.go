package models

import "time"

type Notification struct {
	NotificationID int       `json:"notification_id" db:"notification_id"` // Primary Key
	UserID         int       `json:"user_id" db:"user_id"`                 // Foreign Key to User
	RecipientID    int       `json:"recipient_id" db:"recipient_id"`       // Foreign Key to Recipient
	ThresholdID    int       `json:"threshold_id" db:"threshold_id"`       // Foreign Key to Threshold
	SentAt         time.Time `json:"sent_at" db:"sent_at"`                 // Time sent
	UserMsg        string    `json:"user_msg" db:"user_msg"`               // Message to user
	RecipientMsg   string    `json:"recipient_msg" db:"recipient_msg"`     // Message to recipient
}
