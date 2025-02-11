package services

import (
	"bytes"
	"log"
	"testing"

	"megga-backend/internal/models"
	"megga-backend/internal/services"
)

func TestSendNotifications(t *testing.T) {
	threshold := models.Threshold{
		ThresholdID:    1,
		UserID:         1,
		ThresholdValue: 10.0,
		NotifyUser:     true,
	}

	// 🎯 Mock recipients
	recipients := []models.Recipient{
		{RecipientID: 1, Email: "test@example.com", FirstName: "Test", LastName: "User"},
	}

	// 🎯 Mock user email
	userEmail := "user@example.com"

	// 🎯 Capture logs
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(nil)

	// 🎯 Run function with required arguments
	services.SendNotifications(threshold, "Milk, Fresh, Low Fat", 12.0, recipients, userEmail)

	// 🛠 Print the actual logs for debugging
	actualLogs := logBuffer.String()
	t.Logf("🔍 Captured Logs:\n%s", actualLogs)

	// 🎯 Check logs
	expectedLog := "📧 [MOCK EMAIL] To: test@example.com"
	if !bytes.Contains([]byte(actualLogs), []byte(expectedLog)) {
		t.Errorf("❌ Expected log: %s", expectedLog)
	}

	expectedUserLog := "📧 [MOCK EMAIL] To: user@example.com"
	if !bytes.Contains([]byte(actualLogs), []byte(expectedUserLog)) {
		t.Errorf("❌ Expected user email log: %s", expectedUserLog)
	}
}
