package models

type Recipient struct {
	RecipientID int    `json:"recipient_id" db:"recipient_id"` // Primary Key
	Email       string `json:"email" db:"email"`               // Email address
	FirstName   string `json:"first_name" db:"first_name"`     // First name
	LastName    string `json:"last_name" db:"last_name"`       // Last name
	Designation string `json:"designation" db:"designation"`   // E.g., "Representative"
}
