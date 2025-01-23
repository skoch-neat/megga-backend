package models

import "time"

type Data struct {
	DataID              int       `json:"data_id" db:"data_id"`                     // Primary Key
	Name                string    `json:"name" db:"name"`                               // Name of the item
	Type                string    `json:"type" db:"type"`                               // Type (good/indicator)
	Unit                string    `json:"unit" db:"unit"`                               // Unit of measurement
	PreviousValue       float64   `json:"previous_value" db:"previous_value"`           // Last value
	UpdatedValue        float64   `json:"updated_value" db:"updated_value"`             // Current value
	LastUpdated         time.Time `json:"last_updated" db:"last_updated"`               // When updated
	UpdateIntervalInDays int      `json:"update_interval_in_days" db:"update_interval"` // API update frequency
}
