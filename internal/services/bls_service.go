package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"megga-backend/internal/config"
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
			} `json:"data"`
		} `json:"series"`
	} `json:"Results"`
}

// FetchLatestBLSData requests the latest BLS data for all series
func FetchLatestBLSData() (map[string]float64, error) {
	// Ensure API key is set
	if config.BLS_API_KEY == "" {
		log.Println("❌ Error: BLS_API_KEY is missing. Ensure it's set in the .env file.")
		return nil, fmt.Errorf("missing BLS API key")
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"seriesid":       config.BLS_SERIES_IDS, // ✅ Ensure this is a list
		"catalog":        true,
		"calculations":   true,
		"annualaverage":  false,
		"aspects":        false,
		"registrationkey": config.BLS_API_KEY, // ✅ Required for registered users
	}

	// Log the payload to verify before sending
	payloadJSON, _ := json.Marshal(payload)
	log.Printf("📤 Sending request to BLS API with payload: %s", payloadJSON)

	// Send API request
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", BLS_API_URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to BLS API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Log response body for debugging
	log.Printf("📥 BLS API response: %s", body)

	// Parse JSON response
	var blsResponse BLSResponse
	if err := json.Unmarshal(body, &blsResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Check API response status
	if blsResponse.Status != "REQUEST_SUCCEEDED" {
		return nil, fmt.Errorf("BLS API request failed: %s", blsResponse.Status)
	}

	// Extract latest data
	dataMap := make(map[string]float64)
	for _, series := range blsResponse.Results.Series {
		if len(series.Data) > 0 {
			latest := series.Data[0]
			var value float64
			fmt.Sscanf(latest.Value, "%f", &value)
			dataMap[series.SeriesID] = value
		}
	}

	log.Printf("✅ Fetched BLS Data: %+v", dataMap)
	return dataMap, nil
}
