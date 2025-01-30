package handlers

import (
	"context"
	"encoding/json"
	"megga-backend/models"
	"megga-backend/services/database"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

// CreateThresholdRecipient inserts a new threshold recipient
func CreateThresholdRecipient(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var recipient models.ThresholdRecipient

	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if recipient.ThresholdID == 0 || recipient.RecipientID == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO threshold_recipients (threshold_id, recipient_id, is_user)
		VALUES ($1, $2, $3)
		RETURNING threshold_id, recipient_id, is_user
	`
	err := db.QueryRow(context.Background(), query, recipient.ThresholdID, recipient.RecipientID, recipient.IsUser).
		Scan(&recipient.ThresholdID, &recipient.RecipientID, &recipient.IsUser)

	if err != nil {
		http.Error(w, "Database insert error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "Threshold recipient created successfully",
		"thresholdRecipient": recipient,
	})
}

// GetThresholdRecipients retrieves all threshold recipients
func GetThresholdRecipients(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var recipients []models.ThresholdRecipient

	query := `SELECT threshold_id, recipient_id, is_user FROM threshold_recipients`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var recipient models.ThresholdRecipient
		if err := rows.Scan(&recipient.ThresholdID, &recipient.RecipientID, &recipient.IsUser); err != nil {
			http.Error(w, "Error scanning recipients", http.StatusInternalServerError)
			return
		}
		recipients = append(recipients, recipient)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipients)
}

// GetThresholdRecipientByID retrieves a single recipient by ThresholdID and RecipientID
func GetThresholdRecipientByID(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	thresholdID, err1 := strconv.Atoi(vars["threshold_id"])
	recipientID, err2 := strconv.Atoi(vars["recipient_id"])

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid threshold or recipient ID", http.StatusBadRequest)
		return
	}

	var recipient models.ThresholdRecipient
	query := `
		SELECT threshold_id, recipient_id, is_user
		FROM threshold_recipients WHERE threshold_id = $1 AND recipient_id = $2
	`
	err := db.QueryRow(context.Background(), query, thresholdID, recipientID).Scan(&recipient.ThresholdID, &recipient.RecipientID, &recipient.IsUser)

	if err == pgx.ErrNoRows {
		http.Error(w, "Threshold recipient not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipient)
}

// UpdateThresholdRecipient updates a recipient's `is_user` status
func UpdateThresholdRecipient(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	thresholdID, err1 := strconv.Atoi(vars["threshold_id"])
	recipientID, err2 := strconv.Atoi(vars["recipient_id"])

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid threshold or recipient ID", http.StatusBadRequest)
		return
	}

	var recipient models.ThresholdRecipient
	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE threshold_recipients
		SET is_user = $1
		WHERE threshold_id = $2 AND recipient_id = $3
	`
	_, err := db.Exec(context.Background(), query, recipient.IsUser, thresholdID, recipientID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Threshold recipient updated successfully"})
}

// DeleteThresholdRecipient removes a recipient from a threshold
func DeleteThresholdRecipient(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	thresholdID, err1 := strconv.Atoi(vars["threshold_id"])
	recipientID, err2 := strconv.Atoi(vars["recipient_id"])

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid threshold or recipient ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM threshold_recipients WHERE threshold_id = $1 AND recipient_id = $2"
	res, err := db.Exec(context.Background(), query, thresholdID, recipientID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Threshold recipient not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Threshold recipient deleted successfully"})
}

// RegisterThresholdRecipientRoutes registers all routes
func RegisterThresholdRecipientRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/threshold_recipients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateThresholdRecipient(w, r, db)
		} else if r.Method == "GET" {
			GetThresholdRecipients(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "GET")

	router.HandleFunc("/threshold_recipients/{threshold_id:[0-9]+}/{recipient_id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetThresholdRecipientByID(w, r, db)
		} else if r.Method == "PUT" {
			UpdateThresholdRecipient(w, r, db)
		} else if r.Method == "DELETE" {
			DeleteThresholdRecipient(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT", "DELETE")
}
