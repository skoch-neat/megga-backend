package models

import "time"

type DataHistory struct {
	HistoryID   int       `json:"history_id" db:"history_id"`
	DataID      int       `json:"data_id" db:"data_id"`
	Year        string    `json:"year" db:"year"`
	Period      string    `json:"period" db:"period"`
	PeriodName  string    `json:"period_name" db:"period_name"`
	Value       float64   `json:"value" db:"value"`
	RecordedAt  time.Time `json:"recorded_at" db:"recorded_at"`
}