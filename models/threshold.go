package models

import "time"

type Threshold struct {
	ThresholdID    int       `json:"threshold_id" db:"threshold_id"`       // Primary Key
	UserPoolID     int       `json:"user_pool_id" db:"user_pool_id"`       // Foreign Key to UserPool
	DataID         int       `json:"data_id" db:"data_id"`                 // Foreign Key to Data
	ThresholdValue float64   `json:"threshold_value" db:"threshold_value"` // Percentage change
	CreatedAt      time.Time `json:"created_at" db:"created_at"`           // Date created
	NotifyUser     bool      `json:"notify_user" db:"notify_user"`         // Notify users
}
