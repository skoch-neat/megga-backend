package handlers

import (
	"context"
	"encoding/json"
	"log"
	"megga-backend/internal/models"
	"megga-backend/internal/database"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
)

func GetUserByEmail(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	email, err := url.QueryUnescape(vars["email"])
	if err != nil {
		http.Error(w, "Invalid email encoding", http.StatusBadRequest)
		return
	}

	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	var user models.User
	query := "SELECT user_id, email, first_name, last_name FROM users WHERE LOWER(email) = LOWER($1)"
	err = db.QueryRow(context.Background(), query, email).Scan(&user.UserID, &user.Email, &user.FirstName, &user.LastName)

	if err == pgx.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("❌ Database error fetching user (%s): %v", email, err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func CreateUser(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Email == "" || user.FirstName == "" || user.LastName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var existingID int
	query := "SELECT user_id FROM users WHERE email = $1"
	err := db.QueryRow(context.Background(), query, user.Email).Scan(&existingID)
	if err == pgx.ErrNoRows {
		query = `INSERT INTO users (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING user_id, email`
		if err := db.QueryRow(context.Background(), query, user.Email, user.FirstName, user.LastName).Scan(&user.UserID, &user.Email); err != nil {
			http.Error(w, "Database insert error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User created successfully",
			"user":    user,
		})
		return
	} else if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}

	user.UserID = existingID
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User already exists",
		"user":    user,
	})
}

func GetThresholdsForUser(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid or missing user ID", http.StatusBadRequest)
		return
	}

	log.Printf("🔍 Fetching thresholds for user_id: %d", userID)

	type ThresholdWithRecipients struct {
		ThresholdID    int     `json:"threshold_id"`
		DataID         int     `json:"data_id"`
		Name           string  `json:"name"`
		ThresholdValue float64 `json:"threshold_value"`
		NotifyUser     bool    `json:"notify_user"`
		Recipients     []int64 `json:"recipients"`
	}

	var thresholds []ThresholdWithRecipients

	query := `
		SELECT t.threshold_id, t.data_id, d.name, t.threshold_value, t.notify_user, 
		       COALESCE(ARRAY_AGG(tr.recipient_id) FILTER (WHERE tr.recipient_id IS NOT NULL), ARRAY[]::BIGINT[]) AS recipients
		FROM thresholds t
		JOIN data d ON t.data_id = d.data_id
		LEFT JOIN threshold_recipients tr ON t.threshold_id = tr.threshold_id
		WHERE t.user_id = $1
		GROUP BY t.threshold_id, d.name
	`
	log.Printf("🔍 Executing query: %s", query)

	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		log.Printf("❌ Database Query Error: %v", err)
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		log.Println("✅ No thresholds found, returning empty array.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ThresholdWithRecipients{})
		return
	}

	for {
		var threshold ThresholdWithRecipients
		log.Println("🔍 Scanning row data...")

		if err := rows.Scan(&threshold.ThresholdID, &threshold.DataID, &threshold.Name, &threshold.ThresholdValue, &threshold.NotifyUser, pq.Array(&threshold.Recipients)); err != nil {
			log.Printf("❌ Error Scanning Data: %v", err)
			http.Error(w, "Error scanning data", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ Scanned Threshold: %+v", threshold)
		thresholds = append(thresholds, threshold)

		if !rows.Next() {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thresholds)
}

func DeleteAllThresholdsForUser(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid or missing user ID", http.StatusBadRequest)
		return
	}

	log.Printf("🗑 Deleting all thresholds for user_id: %d...", userID)

	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Printf("❌ Error Starting Transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "DELETE FROM threshold_recipients WHERE threshold_id IN (SELECT threshold_id FROM thresholds WHERE user_id = $1)", userID)
	if err != nil {
		log.Printf("❌ Error deleting threshold recipients: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM thresholds WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("❌ Error deleting thresholds: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(context.Background()); err != nil {
		log.Printf("❌ Error Committing Transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ All thresholds deleted for user_id: %d", userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All thresholds deleted successfully"})
}

func RegisterUserRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateUser(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST")

	router.HandleFunc("/users/{email}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetUserByEmail(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET")

	router.HandleFunc("/users/{userId}/thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetThresholdsForUser(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET")

	router.HandleFunc("/users/{userId}/thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			DeleteAllThresholdsForUser(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("DELETE")
}
