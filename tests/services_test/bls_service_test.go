package services_test

import (
	"testing"

	"megga-backend/internal/services"

	"github.com/pashagolub/pgxmock"
)

func TestFetchBLSData_Timeout(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mock.Close()

	err = services.FetchLatestBLSData(mock)
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
}

func TestParseBLSResponse_MalformedJSON(t *testing.T) {
	malformedJSON := `{"status": "REQUEST_SUCCEEDED", "Results": { "series": [`
	_, err := services.ParseBLSResponse([]byte(malformedJSON))

	if err == nil {
		t.Errorf("Expected JSON parsing error, got nil")
	}
}

