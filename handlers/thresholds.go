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
)

// CreateThreshold handles creating a new threshold
func CreateThreshold(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var threshold models.Threshold

	if err := json.NewDecoder(r.Body).Decode(&threshold); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if threshold.DataID == 0 || threshold.ThresholdValue == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO thresholds (data_id, threshold_value)
		VALUES ($1, $2)
		RETURNING threshold_id
	`
	err := db.QueryRow(context.Background(), query, threshold.DataID, threshold.ThresholdValue).Scan(&threshold.ThresholdID)
	if err != nil {
		log.Printf("Error inserting threshold: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Threshold created successfully",
		"threshold": threshold,
	})
}

// GetThresholds retrieves all thresholds
func GetThresholds(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var thresholds []models.Threshold

	rows, err := db.Query(context.Background(), "SELECT threshold_id, data_id, threshold_value FROM thresholds")
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var threshold models.Threshold
		if err := rows.Scan(&threshold.ThresholdID, &threshold.DataID, &threshold.ThresholdValue); err != nil {
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

	// Read body into a buffer to allow multiple reads
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate JSON field presence
	if !jsonFieldExists(body, "threshold_value") {
		http.Error(w, "Missing required field: threshold_value", http.StatusBadRequest)
		return
	}

	// Decode JSON into struct
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

// DeleteThreshold removes a threshold by ID
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

	// Execute the delete query
	query := "DELETE FROM thresholds WHERE threshold_id = $1"
	res, err := db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Ensure at least one row was deleted
	if res.RowsAffected() == 0 {
		http.Error(w, "Threshold not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Threshold deleted successfully"})
}

// Check if a JSON field exists in the request
func jsonFieldExists(body []byte, field string) bool {
	var requestData map[string]interface{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		return false
	}
	_, exists := requestData[field]
	return exists
}

// RegisterThresholdRoutes registers threshold routes
func RegisterThresholdRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateThreshold(w, r, db)
		} else if r.Method == "GET" {
			GetThresholds(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "GET")

	router.HandleFunc("/thresholds/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			UpdateThreshold(w, r, db)
		} else if r.Method == "DELETE" {
			DeleteThreshold(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("PUT", "DELETE")
}
