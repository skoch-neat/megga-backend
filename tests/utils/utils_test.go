package utils

import (
	"testing"
	"megga-backend/internal/utils"
)

func TestCalculatePercentChange(t *testing.T) {
	tests := []struct {
		name           string
		previousValue  float64
		latestValue    float64
		expectedChange float64
	}{
		{"Increase", 5.0, 5.5, 10.0},
		{"Decrease", 5.0, 4.5, -10.0},
		{"No Change", 5.0, 5.0, 0.0},
		{"Zero Previous Value", 0.0, 5.0, 0.0}, // Avoid division by zero
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CalculatePercentChange(tt.previousValue, tt.latestValue)
			if result != tt.expectedChange {
				t.Errorf("expected %.2f%%, got %.2f%%", tt.expectedChange, result)
			}
		})
	}
}
