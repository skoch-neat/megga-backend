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
var BLS_API_KEY = getBLSAPIKey()

func getBLSAPIURL() string {
	var blsURL = os.Getenv("BLS_API_URL")
	if config.IsDevelopmentMode() {
		if blsURL == "" {
			log.Fatal("BLS_API_URL is not set, check your environment variables")
		} else {
			log.Println("BLS_API_URL is set: ", blsURL)
		}
	}
	return blsURL
}

func getBLSAPIKey() string {
	var blsAPIKey = os.Getenv("BLS_API_KEY")
	if config.IsDevelopmentMode() {
		if blsAPIKey == "" {
			log.Fatal("BLS_API_KEY is not set, check your environment variables")
		} else {
			log.Println("BLS_API_KEY is set: ", blsAPIKey)
		}
	}
	return blsAPIKey
}

// Structs to parse API response
type BLSResponse struct {
	Status  string `json:"status"`
	Results struct {
		Series []struct {
			SeriesID string `json:"seriesID"`
			Data     []struct {
				Year   string `json:"year"`
				Period string `json:"period"`
				Value  string `json:"value"`
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

func FetchLatestBLSData(db database.DBQuerier) error {
	BLS_API_URL = getBLSAPIURL()
	log.Println("🌐 Fetching latest BLS data...")

	seriesIDs := make([]string, 0, len(config.BLS_SERIES_INFO))
	for seriesID := range config.BLS_SERIES_INFO {
		seriesIDs = append(seriesIDs, seriesID)
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if config.IsDevelopmentMode() {
		log.Printf("📥 BLS API response: %s", body)
	}

	blsData, err := ParseBLSResponse(body)
	if err != nil {
		return fmt.Errorf("error parsing BLS response: %w", err)
	}

	err = SaveBLSData(db, blsData)
	if err != nil {
		return fmt.Errorf("error saving BLS data: %w", err)
	}

	log.Println("✅ BLS data saved successfully.")

	log.Println("🔍 Checking thresholds against updated BLS data...")

	return SaveBLSData(db, blsData)
}
