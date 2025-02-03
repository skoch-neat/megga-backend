package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"megga-backend/models"
	"megga-backend/services/database"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func CreateThreshold(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var request struct {
		UserPoolID     int     `json:"user_pool_id"`
		DataID         int     `json:"data_id"`
		ThresholdValue float64 `json:"threshold_value"`
		NotifyUsers    bool    `json:"notify_user"`
		Recipients     []int   `json:"recipients"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.UserPoolID == 0 || request.DataID == 0 || request.ThresholdValue == 0 || len(request.Recipients) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var existingID int
	checkQuery := "SELECT threshold_id FROM thresholds WHERE user_pool_id = $1 AND data_id = $2"
	err := db.QueryRow(context.Background(), checkQuery, request.UserPoolID, request.DataID).Scan(&existingID)

	if err == nil {
		http.Error(w, "Threshold already exists for this user and data point", http.StatusConflict)
		return
	}

	var thresholdID int
	insertThresholdQuery := `
		INSERT INTO thresholds (user_pool_id, data_id, threshold_value, created_at, notify_user)
		VALUES ($1, $2, $3, NOW(), $4)
		RETURNING threshold_id
	`
	err = db.QueryRow(context.Background(), insertThresholdQuery, request.UserPoolID, request.DataID, request.ThresholdValue, request.NotifyUsers).Scan(&thresholdID)

	if err != nil {
		log.Printf("Error inserting threshold: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	for _, recipientID := range request.Recipients {
		_, err := db.Exec(context.Background(), "INSERT INTO threshold_recipients (threshold_id, recipient_id) VALUES ($1, $2)", thresholdID, recipientID)
		if err != nil {
			log.Printf("Error inserting recipient for threshold: %v", err)
			http.Error(w, "Error associating recipients", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Threshold created successfully",
		"threshold_id": thresholdID,
	})
}

func GetThresholds(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid or missing user ID", http.StatusBadRequest)
		return
	}

	type ThresholdWithRecipients struct {
		ThresholdID    int     `json:"threshold_id"`
		DataID         int     `json:"data_id"`
		ThresholdValue float64 `json:"threshold_value"`
		NotifyUsers    bool    `json:"notify_user"`
		Recipients     []int   `json:"recipients"`
	}

	var thresholds []ThresholdWithRecipients

	query := `
		SELECT t.threshold_id, t.data_id, t.threshold_value, t.notify_user, 
		       COALESCE(ARRAY_AGG(tr.recipient_id) FILTER (WHERE tr.recipient_id IS NOT NULL), '{}') AS recipients
		FROM thresholds t
		LEFT JOIN threshold_recipients tr ON t.threshold_id = tr.threshold_id
		WHERE t.user_pool_id = $1
		GROUP BY t.threshold_id
	`
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var threshold ThresholdWithRecipients
		if err := rows.Scan(&threshold.ThresholdID, &threshold.DataID, &threshold.ThresholdValue, &threshold.NotifyUsers, pq.Array(&threshold.Recipients)); err != nil {
			http.Error(w, "Error scanning data", http.StatusInternalServerError)
			return
		}
		thresholds = append(thresholds, threshold)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thresholds)
}

func UpdateThreshold(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid threshold ID", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !jsonFieldExists(body, "threshold_value") {
		http.Error(w, "Missing required field: threshold_value", http.StatusBadRequest)
		return
	}

	var threshold models.Threshold
	if err := json.Unmarshal(body, &threshold); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE thresholds
		SET threshold_value = $1
		WHERE threshold_id = $2
		RETURNING threshold_id, data_id, threshold_value
	`
	err = db.QueryRow(context.Background(), query, threshold.ThresholdValue, id).
		Scan(&threshold.ThresholdID, &threshold.DataID, &threshold.ThresholdValue)
	if err != nil {
		log.Printf("Error updating threshold: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Threshold updated successfully",
		"threshold": threshold,
	})
}

func DeleteThreshold(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists || idStr == "" {
		http.Error(w, "Missing threshold ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid threshold ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM thresholds WHERE threshold_id = $1"
	res, err := db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if res.RowsAffected() == 0 {
		http.Error(w, "Threshold not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Threshold deleted successfully"})
}

func jsonFieldExists(body []byte, field string) bool {
	var requestData map[string]interface{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		return false
	}
	_, exists := requestData[field]
	return exists
}

func RegisterThresholdRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/thresholds/{userId}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetThresholds(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET")

	router.HandleFunc("/thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateThreshold(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST")
}
