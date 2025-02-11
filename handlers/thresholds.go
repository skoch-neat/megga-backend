package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"

	"github.com/lib/pq"
)

func CreateThreshold(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var request models.Threshold

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("🔍 Raw Request Body: %s", string(body))

	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("❌ JSON Unmarshal Error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("✅ Decoded Request: %+v", request)

	if request.UserID == 0 || request.DataID == 0 || request.ThresholdValue == 0 || len(request.Recipients) == 0 {
		log.Printf("❌ Missing Required Fields: UserID=%d, DataID=%d, ThresholdValue=%f, Recipients=%v",
			request.UserID, request.DataID, request.ThresholdValue, request.Recipients)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	log.Printf("✅ Preparing to Insert: UserID=%d, DataID=%d, ThresholdValue=%.2f, NotifyUser=%t, Recipients=%v",
		request.UserID, request.DataID, request.ThresholdValue, request.NotifyUser, request.Recipients)

	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Printf("❌ Error Starting Transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	var thresholdID int
	query := `INSERT INTO thresholds (user_id, data_id, threshold_value, notify_user, created_at)
	          VALUES ($1, $2, $3, $4, NOW()) RETURNING threshold_id`
	err = tx.QueryRow(context.Background(), query, request.UserID, request.DataID, request.ThresholdValue, request.NotifyUser).
		Scan(&thresholdID)

	if err != nil {
		log.Printf("❌ Error Inserting Threshold: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Inserted Threshold: ID=%d", thresholdID)

	for _, recipientID := range request.Recipients {
		log.Printf("🔍 Attempting to insert recipient: ThresholdID=%d, RecipientID=%d", thresholdID, recipientID)

		_, err := tx.Exec(context.Background(),
			"INSERT INTO threshold_recipients (threshold_id, recipient_id) VALUES ($1, $2)",
			thresholdID, recipientID)

		if err != nil {
			log.Printf("❌ Error inserting recipient for threshold: %v", err)
			http.Error(w, "Error associating recipients", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ Successfully inserted: ThresholdID=%d, RecipientID=%d", thresholdID, recipientID)
	}

	if err := tx.Commit(context.Background()); err != nil {
		log.Printf("❌ Error Committing Transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Println("✅ Threshold Created Successfully")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Threshold created successfully",
		"threshold_id": thresholdID,
	})
}

func GetThresholdById(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	log.Println("🔍 Fetching threshold by ID...")
	vars := mux.Vars(r)
	thresholdID, err := strconv.Atoi(vars["id"])
	if err != nil || thresholdID <= 0 {
		http.Error(w, "Invalid or missing threshold ID", http.StatusBadRequest)
		return
	}

	log.Printf("🔍 Fetching details for threshold ID: %d", thresholdID)

	type ThresholdWithRecipients struct {
		ThresholdID    int     `json:"threshold_id"`
		DataID         int     `json:"data_id"`
		Name           string  `json:"name"`
		ThresholdValue float64 `json:"threshold_value"`
		NotifyUser     bool    `json:"notify_user"`
		Recipients     []int64 `json:"recipients"`
	}

	var threshold ThresholdWithRecipients

	query := `
		SELECT t.threshold_id, t.data_id, d.name, t.threshold_value, t.notify_user, 
		       COALESCE(ARRAY_AGG(tr.recipient_id) FILTER (WHERE tr.recipient_id IS NOT NULL), ARRAY[]::BIGINT[]) AS recipients
		FROM thresholds t
		JOIN data d ON t.data_id = d.data_id
		LEFT JOIN threshold_recipients tr ON t.threshold_id = tr.threshold_id
		WHERE t.threshold_id = $1
		GROUP BY t.threshold_id, d.name
	`
	log.Printf("🔍 Executing query: %s with threshold_id=%d", query, thresholdID)

	log.Printf("🔍 Executing query: %s with threshold_id=%d", query, thresholdID)

	err = db.QueryRow(context.Background(), query, thresholdID).Scan(
		&threshold.ThresholdID, &threshold.DataID, &threshold.Name,
		&threshold.ThresholdValue, &threshold.NotifyUser, pq.Array(&threshold.Recipients),
	)

	if err == pgx.ErrNoRows {
		log.Printf("✅ No threshold found for ID %d", thresholdID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	} else if err != nil {
		log.Printf("❌ Database Query Error: %v", err)
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Retrieved Threshold: %+v", threshold)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(threshold)
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

	var threshold models.Threshold
	if err := json.Unmarshal(body, &threshold); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("✏️ Updating Threshold ID %d: %+v", id, threshold)

	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		http.Error(w, "Database transaction error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	query := `
		UPDATE thresholds
		SET threshold_value = $1, notify_user = $2
		WHERE threshold_id = $3
		RETURNING threshold_id
	`
	err = tx.QueryRow(context.Background(), query, threshold.ThresholdValue, threshold.NotifyUser, id).Scan(&threshold.ThresholdID)
	if err != nil {
		log.Printf("❌ Error updating threshold: %v", err)
		http.Error(w, "Database update error", http.StatusInternalServerError)
		return
	}

	var existingRecipientIDs []int
	getRecipientsQuery := `SELECT recipient_id FROM threshold_recipients WHERE threshold_id = $1`
	rows, err := tx.Query(context.Background(), getRecipientsQuery, id)
	if err != nil {
		http.Error(w, "Error fetching existing recipients", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var recipientID int
		if err := rows.Scan(&recipientID); err != nil {
			http.Error(w, "Error reading existing recipients", http.StatusInternalServerError)
			return
		}
		existingRecipientIDs = append(existingRecipientIDs, recipientID)
	}

	newRecipientSet := make(map[int]bool)
	for _, r := range threshold.Recipients {
		newRecipientSet[r] = true
	}

	for _, existingID := range existingRecipientIDs {
		if !newRecipientSet[existingID] {
			deleteQuery := `DELETE FROM threshold_recipients WHERE threshold_id = $1 AND recipient_id = $2`
			_, err := tx.Exec(context.Background(), deleteQuery, id, existingID)
			if err != nil {
				http.Error(w, "Error removing old recipients", http.StatusInternalServerError)
				return
			}
			log.Printf("❌ Removed recipient %d from threshold %d", existingID, id)
		}
	}

	existingRecipientMap := make(map[int]bool)
	for _, existingID := range existingRecipientIDs {
		existingRecipientMap[existingID] = true
	}

	for _, newID := range threshold.Recipients {
		if !existingRecipientMap[newID] {
			insertQuery := `INSERT INTO threshold_recipients (threshold_id, recipient_id) VALUES ($1, $2)`
			_, err := tx.Exec(context.Background(), insertQuery, id, newID)
			if err != nil {
				http.Error(w, "Error adding new recipients", http.StatusInternalServerError)
				return
			}
			log.Printf("✅ Added recipient %d to threshold %d", newID, id)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Successfully updated threshold ID %d", id)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Threshold updated successfully",
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

func RegisterThresholdRoutes(router *mux.Router, db database.DBQuerier) {
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
}
