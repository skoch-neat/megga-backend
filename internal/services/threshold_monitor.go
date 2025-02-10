package services

import (
	"context"
	"log"
	"megga-backend/internal/config"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"megga-backend/internal/utils"
)

func CheckThresholdsAndNotify(db database.DBQuerier) {
	MonitorThresholds(db)
}

func MonitorThresholds(db database.DBQuerier) {
	log.Println("🔍 Running scheduled threshold checks...")

	thresholds, err := fetchAllThresholds(db)
	if err != nil {
		log.Printf("❌ Failed to fetch thresholds: %v", err)
		return
	}

	for _, threshold := range thresholds {
		latestValue, err := fetchLatestDataValue(db, threshold.DataID)
		if err != nil {
			log.Printf("❌ Error fetching latest value for data ID %d: %v", threshold.DataID, err)
			continue
		}

		dataIDStr := utils.ConvertIntToString(threshold.DataID)

		if _, exists := config.BLS_SERIES_INFO[dataIDStr]; !exists {
			continue
		}

		percentChange := utils.CalculatePercentChange(threshold.ThresholdValue, latestValue)
		if percentChange >= threshold.ThresholdValue {
			log.Printf("⚠️ Threshold exceeded for %d (Data ID: %d) - Triggering notifications", threshold.ThresholdID, threshold.DataID)

			dataName, err := utils.FetchDataName(db, threshold.DataID)
			if err != nil {
				log.Printf("❌ Failed to fetch data name for Data ID %d", threshold.DataID)
				continue
			}

			SendNotifications(db, threshold, dataName, percentChange)
		}
	}
}

func fetchAllThresholds(db database.DBQuerier) ([]models.Threshold, error) {
	rows, err := db.Query(context.Background(), "SELECT threshold_id, user_id, data_id, threshold_value, notify_user FROM thresholds")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var thresholds []models.Threshold
	for rows.Next() {
		var threshold models.Threshold
		if err := rows.Scan(&threshold.ThresholdID, &threshold.UserID, &threshold.DataID, &threshold.ThresholdValue, &threshold.NotifyUser); err != nil {
			return nil, err
		}
		thresholds = append(thresholds, threshold)
	}
	return thresholds, nil
}

func fetchLatestDataValue(db database.DBQuerier, dataID int) (float64, error) {
	var latestValue float64
	err := db.QueryRow(context.Background(), "SELECT latest_value FROM data WHERE data_id = $1", dataID).Scan(&latestValue)
	if err != nil {
		return 0, err
	}
	return latestValue, nil
}
