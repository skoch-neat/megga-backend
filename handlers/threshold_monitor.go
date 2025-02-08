package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var ApprovedBlsSeriesIDs = map[string]bool{
	"APU0000708111": true, // Eggs, grade A, large, per doz.
	"APU0000702111": true, // Bread, white, pan, per lb.
	"APU0000709213": true, // Milk, fresh, low fat, per gal.
	"APU0000FF1101": true, // Chicken breast, boneless, per lb.
	"APU0000704111": true, // Bacon, sliced, per lb.
	"APU0000711111": true, // Apples, Red Delicious, per lb.
	"APU0000711311": true, // Oranges, Navel, per lb.
	"APU00007471A":  true, // Gasoline, all types, per gal.
	"LEU0252881600": true, // Median usual weekly earnings
}

func CheckThresholdsHandler(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	log.Println("🔍 Manually triggering threshold check...")

	CheckThresholdsAndNotify(db)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Thresholds checked successfully",
	})
}

func CheckThresholdsAndNotify(db database.DBQuerier) {
	log.Println("🔍 Checking thresholds...")

	dataQuery := `SELECT data_id, name, series_id, latest_value, previous_value FROM data`
	rows, err := db.Query(context.Background(), dataQuery)
	if err != nil {
		log.Printf("❌ Database query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var data models.Data
		if err := rows.Scan(&data.DataID, &data.Name, &data.SeriesID, &data.LatestValue, &data.PreviousValue); err != nil {
			log.Printf("❌ Error scanning data: %v", err)
			continue
		}

		if !ApprovedBlsSeriesIDs[data.SeriesID] {
			continue
		}

		if data.PreviousValue == 0 {
			log.Printf("⚠️ Skipping %s (DataID: %d) due to zero previous value", data.Name, data.DataID)
			continue
		}
		percentageChange := ((data.LatestValue - data.PreviousValue) / data.PreviousValue) * 100

		thresholdQuery := `
			SELECT threshold_id, user_id, threshold_value, notify_user
			FROM thresholds WHERE data_id = $1
		`
		thresholdRows, err := db.Query(context.Background(), thresholdQuery, data.DataID)
		if err != nil {
			log.Printf("❌ Error fetching thresholds: %v", err)
			continue
		}
		defer thresholdRows.Close()

		for thresholdRows.Next() {
			var threshold models.Threshold
			if err := thresholdRows.Scan(&threshold.ThresholdID, &threshold.UserID, &threshold.ThresholdValue, &threshold.NotifyUser); err != nil {
				log.Printf("❌ Error scanning threshold: %v", err)
				continue
			}

			if percentageChange >= threshold.ThresholdValue || percentageChange <= -threshold.ThresholdValue {
				log.Printf("🚨 Threshold Exceeded for %s (DataID: %d) | Change: %.2f%% | Threshold: %.2f%%",
					data.Name, data.DataID, percentageChange, threshold.ThresholdValue)

				sendNotifications(db, threshold, data, percentageChange)
			}
		}
	}
}

func sendNotifications(db database.DBQuerier, threshold models.Threshold, data models.Data, percentageChange float64) {
	log.Printf("📩 Fetching recipients for ThresholdID: %d", threshold.ThresholdID)

	recipientQuery := `
		SELECT r.recipient_id, r.email, r.first_name, r.last_name 
		FROM threshold_recipients tr
		JOIN recipients r ON tr.recipient_id = r.recipient_id
		WHERE tr.threshold_id = $1
	`
	recipientRows, err := db.Query(context.Background(), recipientQuery, threshold.ThresholdID)
	if err != nil {
		log.Printf("❌ Error fetching recipients: %v", err)
		return
	}
	defer recipientRows.Close()

	subject := fmt.Sprintf("Threshold Alert: %s Exceeded Limit", data.Name)
	message := fmt.Sprintf("The value of %s has changed by %.2f%%, exceeding your set threshold of %.2f%%.",
		data.Name, percentageChange, threshold.ThresholdValue)

	for recipientRows.Next() {
		var recipient models.Recipient
		if err := recipientRows.Scan(&recipient.RecipientID, &recipient.Email, &recipient.FirstName, &recipient.LastName); err != nil {
			log.Printf("❌ Error scanning recipient: %v", err)
			continue
		}

		sendEmail(recipient.Email, subject, message)

		insertNotificationQuery := `
			INSERT INTO notifications (user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg)
			VALUES ($1, $2, $3, NOW(), $4, $5)
		`
		_, err := db.Exec(context.Background(), insertNotificationQuery, threshold.UserID, recipient.RecipientID, threshold.ThresholdID, message, message)
		if err != nil {
			log.Printf("❌ Error saving notification: %v", err)
		}
	}
}

func sendEmail(recipientEmail, subject, message string) {
	testmailURL := "https://api.testmail.app/send"
	apiKey := "your-testmail-api-key"

	requestBody := fmt.Sprintf(`{
		"to": "%s",
		"subject": "%s",
		"body": "%s",
		"from": "noreply@megga.com"
	}`, recipientEmail, subject, message)

	req, err := http.NewRequest("POST", testmailURL, strings.NewReader(requestBody))
	if err != nil {
		log.Printf("❌ Error creating email request: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Error sending email: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("📧 Email sent to %s with status %d", recipientEmail, resp.StatusCode)
}

func RegisterThresholdMonitorRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/thresholds/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetThresholdById(w, r, db)
		} else if r.Method == "PUT" {
			UpdateThreshold(w, r, db)
		} else if r.Method == "DELETE" {
			DeleteThreshold(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT", "DELETE")

	router.HandleFunc("/thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateThreshold(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST")

	router.HandleFunc("/check-thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CheckThresholdsHandler(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST")
}
