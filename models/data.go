package models

import "time"

type Data struct {
	DataID               int       `json:"data_id" db:"data_id"`                         // Primary Key
	Name                 string    `json:"name" db:"name"`                               // Name of the item
	SeriesID             string    `json:"series_id" db:"series_id"`                     // Series ID
	Type                 string    `json:"type" db:"type"`                               // Type (good/indicator)
	Unit                 string    `json:"unit" db:"unit"`                               // Unit of measurement
	PreviousValue        float64   `json:"previous_value" db:"previous_value"`           // Last value
	LatestValue          float64   `json:"latest_value" db:"latest_value"`               // Current value
	LastUpdated          time.Time `json:"last_updated" db:"last_updated"`               // When updated
	UpdateIntervalInDays int       `json:"update_interval_in_days" db:"update_interval"` // API update frequency
}
