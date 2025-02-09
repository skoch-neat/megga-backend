package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"megga-backend/internal/config"
	"megga-backend/internal/database"
)

// BLS API URL
var BLS_API_URL = getBLSAPIURL()

func getBLSAPIURL() string {
	return os.Getenv("BLS_API_URL")
}

// Structs to parse API response
type BLSResponse struct {
	Status  string `json:"status"`
	Results struct {
		Series []struct {
			SeriesID string `json:"seriesID"`
			Data     []struct {
				Year       string `json:"year"`
				Period     string `json:"period"`
				Value      string `json:"value"`
			} `json:"data"`
		} `json:"series"`
	} `json:"Results"`
}

// ParseBLSResponse parses JSON response from BLS API
func ParseBLSResponse(body []byte) (map[string]struct {
	Value  float64
	Year   string
	Period string
}, error) {
	var blsResponse BLSResponse
	if err := json.Unmarshal(body, &blsResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	if blsResponse.Status != "REQUEST_SUCCEEDED" {
		return nil, errors.New("BLS API request failed: " + blsResponse.Status)
	}

	blsData := make(map[string]struct {
		Value  float64
		Year   string
		Period string
	})

	for _, series := range blsResponse.Results.Series {
		if len(series.Data) == 0 {
			continue
		}
		latestEntry := series.Data[0]
		var latestValue float64
		fmt.Sscanf(latestEntry.Value, "%f", &latestValue)
		blsData[series.SeriesID] = struct {
			Value  float64
			Year   string
			Period string
		}{
			Value:  latestValue,
			Year:   latestEntry.Year,
			Period: latestEntry.Period,
		}
	}

	return blsData, nil
}

// FetchLatestBLSData requests the latest BLS data and updates the database
func FetchLatestBLSData(db database.DBQuerier) error {
	log.Println("📡 Fetching latest BLS data...")

	seriesIDs := make([]string, 0, len(config.BLS_SERIES_INFO))
	for seriesID := range config.BLS_SERIES_INFO {
		seriesIDs = append(seriesIDs, seriesID)
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"seriesid":        seriesIDs,
		"latest":          true,
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

	blsData, err := ParseBLSResponse(body)
	if err != nil {
		return fmt.Errorf("error parsing BLS response: %w", err)
	}
	
	return SaveBLSData(db, blsData)
}
