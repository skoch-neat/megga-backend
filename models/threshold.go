package models

import "time"

type Threshold struct {
	ThresholdID    int       `json:"threshold_id" db:"threshold_id"`       // Primary Key
	DataID         int       `json:"data_id" db:"data_id"`                // Foreign Key to Data
	ThresholdValue float64   `json:"threshold_value" db:"threshold_value"`    // Percentage change
	CreatedAt      time.Time `json:"created_at" db:"created_at"`              // Date created
}
