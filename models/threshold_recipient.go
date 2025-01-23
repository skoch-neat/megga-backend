package models

type ThresholdRecipient struct {
	ThresholdID int  `json:"threshold_id" db:"threshold_id"` // Foreign Key to Threshold
	RecipientID int  `json:"recipient_id" db:"recipient_id"` // Foreign Key to Recipient/User
	IsUser      bool `json:"is_user" db:"is_user"`              // true for user, false for others
}
