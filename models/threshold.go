package models

import "time"

type Threshold struct {
	ThresholdID    int       `json:"thresholdId,omitempty" db:"threshold_id"`
	UserID         int       `json:"userId" db:"user_id"`
	DataID         int       `json:"dataId" db:"data_id"`
	ThresholdValue float64   `json:"thresholdValue" db:"threshold_value"`
	CreatedAt      time.Time `json:"createdAt,omitempty" db:"created_at"`
	NotifyUser     bool      `json:"notifyUser" db:"notify_user"`
	Recipients     []int     `json:"recipients,omitempty"`
}
