package models

type User struct {
	UserID    int    `json:"user_id" db:"user_id"` // Primary Key
	Email     string `json:"email" db:"email"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
}
