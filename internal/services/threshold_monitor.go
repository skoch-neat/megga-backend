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
		log.Printf("🔍 Fetching latest value for Data ID: %d", threshold.DataID)
		latestValue, err := fetchLatestDataValue(db, threshold.DataID)
		if err != nil {
			log.Printf("❌ Error fetching latest value for Data ID %d: %v", threshold.DataID, err)
			continue
		}

		seriesID, err := fetchSeriesIDForData(db, threshold.DataID)
		if err != nil {
			log.Printf("❌ Error fetching series ID for Data ID %d: %v", threshold.DataID, err)
			continue
		}

		if _, exists := config.BLS_SERIES_INFO[seriesID]; !exists {
			log.Printf("⚠️ Skipping Data ID %d as its series ID is not in BLS_SERIES_INFO", threshold.DataID)
			continue
		}

		percentChange := utils.CalculatePercentChange(threshold.ThresholdValue, latestValue)
		if percentChange >= threshold.ThresholdValue || percentChange <= -threshold.ThresholdValue {
			log.Printf("⚠️ Threshold exceeded for Threshold ID %d (Data ID: %d) - Triggering notifications", threshold.ThresholdID, threshold.DataID)

			dataName, err := utils.FetchDataName(db, threshold.DataID)
			if err != nil {
				log.Printf("❌ Failed to fetch data name for Data ID %d", threshold.DataID)
				continue
			}

			log.Printf("🔍 Calculated percent change for Data ID %d: %.2f%%", threshold.DataID, percentChange)

			recipients, err := fetchRecipientsForThreshold(db, threshold.ThresholdID)
			if err != nil {
				log.Printf("❌ Error fetching recipients for Threshold ID %d: %v", threshold.ThresholdID, err)
				continue
			}
			userEmail := fetchUserEmail(db, threshold.UserID)

			SendNotifications(threshold, dataName, percentChange, recipients, userEmail)
		}
	}
}

func fetchAllThresholds(db database.DBQuerier) ([]models.Threshold, error) {
	rows, err := db.Query(context.Background(), `
		SELECT threshold_id, user_id, data_id, threshold_value, notify_user 
		FROM thresholds`)
	if err != nil {
		log.Printf("❌ Failed to fetch thresholds: %v", err)
		return nil, err
	}
	defer rows.Close()

	var thresholds []models.Threshold
	for rows.Next() {
		var threshold models.Threshold
		if err := rows.Scan(&threshold.ThresholdID, &threshold.UserID, &threshold.DataID, &threshold.ThresholdValue, &threshold.NotifyUser); err != nil {
			log.Printf("❌ Error scanning threshold row: %v", err)
			return nil, err
		}
		thresholds = append(thresholds, threshold)
	}

	if err := rows.Err(); err != nil {
		log.Printf("❌ Error iterating over threshold rows: %v", err)
		return nil, err
	}

	log.Printf("✅ Fetched %d thresholds successfully.", len(thresholds))
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

func fetchSeriesIDForData(db database.DBQuerier, dataID int) (string, error) {
	var seriesID string
	err := db.QueryRow(context.Background(), "SELECT series_id FROM data WHERE data_id = $1", dataID).Scan(&seriesID)
	if err != nil {
		return "", err
	}
	return seriesID, nil
}
