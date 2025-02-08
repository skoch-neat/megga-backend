package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"megga-backend/internal/config"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"megga-backend/handlers"
)

// BLS API URL
const BLS_API_URL = "https://api.bls.gov/publicAPI/v2/timeseries/data/"

// Structs to parse API response
type BLSResponse struct {
	Status  string `json:"status"`
	Results struct {
		Series []struct {
			SeriesID string `json:"seriesID"`
			Data     []struct {
				Year       string `json:"year"`
				Period     string `json:"period"`
				PeriodName string `json:"periodName"`
				Value      string `json:"value"`
				Latest     string `json:"latest,omitempty"`
			} `json:"data"`
		} `json:"series"`
	} `json:"Results"`
}

// FetchLatestBLSData requests the latest BLS data and updates the database
func FetchLatestBLSData(db database.DBQuerier) error {
	log.Println("📡 Fetching latest BLS data...")

	// Prepare request payload
	payload := map[string]interface{}{
		"seriesid":       config.BLS_SERIES_IDS,
		"startyear":      "2024",
		"endyear":        "2024",
		"registrationkey": config.BLS_API_KEY,
	}

	client := &http.Client{Timeout: 10 * time.Second}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", BLS_API_URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request to BLS API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	log.Printf("📥 BLS API response: %s", body)

	// Parse JSON response
	var blsResponse BLSResponse
	if err := json.Unmarshal(body, &blsResponse); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	if blsResponse.Status != "REQUEST_SUCCEEDED" {
		return fmt.Errorf("BLS API request failed: %s", blsResponse.Status)
	}

	// Process BLS data
	for _, series := range blsResponse.Results.Series {
		latestEntry := findLatestEntry(series.Data)
		if latestEntry == nil {
			continue
		}

		// Convert latest value to float
		var latestValue float64
		fmt.Sscanf(latestEntry.Value, "%f", &latestValue)

		// Check the database for existing values
		var existing models.Data
		err := db.QueryRow(context.Background(), "SELECT data_id, latest_value, previous_value, last_year, last_period FROM data WHERE series_id = $1",
			series.SeriesID).Scan(&existing.DataID, &existing.LatestValue, &existing.PreviousValue, &existing.LastYear, &existing.LastPeriod)

		if err != nil {
			log.Printf("⚠️ No existing record found for series: %s, skipping.", series.SeriesID)
			continue
		}

		// Store this value in history
		_, err = db.Exec(context.Background(),
			"INSERT INTO data_history (data_id, year, period, period_name, value) VALUES ($1, $2, $3, $4, $5)",
			existing.DataID, latestEntry.Year, latestEntry.Period, latestEntry.PeriodName, latestValue,
		)

		if err != nil {
			log.Printf("❌ Error inserting historical data: %v", err)
		}

		// Only update if it's a new period
		if latestEntry.Year > existing.LastYear || (latestEntry.Year == existing.LastYear && latestEntry.Period > existing.LastPeriod) {
			percentageChange := ((latestValue - existing.LatestValue) / existing.LatestValue) * 100

			_, err := db.Exec(context.Background(),
				"UPDATE data SET previous_value = $1, latest_value = $2, last_year = $3, last_period = $4, last_updated = NOW() WHERE data_id = $5",
				existing.LatestValue, latestValue, latestEntry.Year, latestEntry.Period, existing.DataID,
			)
			if err != nil {
				log.Printf("❌ Error updating database: %v", err)
				continue
			}

			log.Printf("✅ Updated %s: %.2f -> %.2f (%.2f%% change)", series.SeriesID, existing.LatestValue, latestValue, percentageChange)
			handlers.CheckThresholdsAndNotify(db)
		}
	}

	return nil
}

// Find the latest value from BLS API response
func findLatestEntry(data []struct {
	Year       string `json:"year"`
	Period     string `json:"period"`
	PeriodName string `json:"periodName"`
	Value      string `json:"value"`
	Latest     string `json:"latest,omitempty"`
}) *struct {
	Year       string `json:"year"`
	Period     string `json:"period"`
	PeriodName string `json:"periodName"`
	Value      string `json:"value"`
	Latest     string `json:"latest,omitempty"`
} {
	for _, entry := range data {
		if entry.Latest == "true" {
			return &entry
		}
	}
	return nil
}
