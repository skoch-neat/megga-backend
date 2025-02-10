package utils

import (
	"context"
	"log"
	"strconv"
	"megga-backend/internal/database"
)

// FetchDataName retrieves the name of a data entry by its ID.
func FetchDataName(db database.DBQuerier, dataID int) (string, error) {
	var name string
	err := db.QueryRow(context.Background(), "SELECT name FROM data WHERE data_id = $1", dataID).Scan(&name)
	if err != nil {
		log.Printf("❌ Failed to fetch data name for DataID %d: %v", dataID, err)
		return "", err
	}
	return name, nil
}

// ConvertIntToString safely converts an int to a string.
func ConvertIntToString(num int) string {
	return strconv.Itoa(num)
}

// CalculatePercentChange computes the percentage change between two values.
func CalculatePercentChange(previousValue, latestValue float64) float64 {
	if previousValue == 0 {
		return 0
	}
	return ((latestValue - previousValue) / previousValue) * 100
}
