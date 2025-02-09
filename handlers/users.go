package handlers

import (
	"bytes"
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

func GetUserByEmail(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	log.Printf("🔍 DEBUG: Handler received request with headers: %+v", r.Header)

	vars := mux.Vars(r)
	urlEmail := vars["email"]
	if urlEmail == "" {
		log.Println("❌ DEBUG: Email parameter missing from URL")
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	// Extract user info from headers instead of context
	email := r.Header.Get("X-User-Email")
	firstName := r.Header.Get("X-User-FirstName")
	lastName := r.Header.Get("X-User-LastName")

	if email == "" || firstName == "" || lastName == "" {
		log.Println("❌ DEBUG: Missing user information in headers")
		http.Error(w, "Unauthorized: Missing user details", http.StatusUnauthorized)
		return
	}

	log.Printf("🔍 DEBUG: Extracted user details: email=%s, firstName=%s, lastName=%s", email, firstName, lastName)

	if urlEmail != email {
		log.Printf("❌ DEBUG: Email mismatch: Requested=%s, JWT=%s", urlEmail, email)
		http.Error(w, "Unauthorized: Email mismatch", http.StatusUnauthorized)
		return
	}

	log.Println("✅ DEBUG: User is authenticated. Fetching from database...")

	// Fetch user from DB
	var user models.User
	query := "SELECT user_id, email, first_name, last_name FROM users WHERE LOWER(email) = LOWER($1)"
	err := db.QueryRow(context.Background(), query, email).Scan(&user.UserID, &user.Email, &user.FirstName, &user.LastName)

	if err == pgx.ErrNoRows {
		log.Printf("⚠️ DEBUG: User not found: %s. Creating new user internally...", email)

		user, err = CreateUserInternal(db, email, firstName, lastName)
		if err != nil {
			log.Printf("❌ DEBUG: Error creating user: %v", err)
			http.Error(w, "Database insert error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	} else if err != nil {
		log.Printf("❌ DEBUG: Database error fetching user (%s): %v", email, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ DEBUG: User found: %+v", user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func CreateUser(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	log.Println("👤 Calling CreateUser...")

	// Read the request body for debugging
	body, _ := io.ReadAll(r.Body)
	log.Printf("📥 Request Body: %s", string(body))
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	type NewUser struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	var newUser NewUser
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		log.Printf("❌ JSON Decode Error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("✅ Decoded User: %+v", newUser)

	if newUser.Email == "" || newUser.FirstName == "" || newUser.LastName == "" {
		log.Printf("❌ Missing required fields: %+v", newUser)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var existingID int
	query := "SELECT user_id FROM users WHERE email = $1"
	log.Printf("🔍 Checking if user exists: %s", newUser.Email)

	err := db.QueryRow(context.Background(), query, newUser.Email).Scan(&existingID)
	if err == nil {
		log.Printf("✅ User already exists: %s", newUser.Email)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User already exists",
			"user": map[string]interface{}{
				"user_id":    existingID,
				"email":      newUser.Email,
				"first_name": newUser.FirstName,
				"last_name":  newUser.LastName,
			},
		})
		return
	}

	if err == pgx.ErrNoRows {
		log.Println("🆕 User does not exist. Proceeding with INSERT...")

		query := `INSERT INTO users (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING user_id, email, first_name, last_name`
		var createdUser models.User
		err := db.QueryRow(context.Background(), query, newUser.Email, newUser.FirstName, newUser.LastName).
			Scan(&createdUser.UserID, &createdUser.Email, &createdUser.FirstName, &createdUser.LastName)

		if err != nil {
			log.Printf("❌ Database INSERT Error: %v", err) // 🔥 THIS will tell us why INSERT fails!
			http.Error(w, "Database insert error", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ User successfully created: %+v", createdUser)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User created successfully",
			"user":    createdUser,
		})
		return
	}

	log.Printf("❌ Unexpected Database Error: %v", err)
	http.Error(w, "Database query error", http.StatusInternalServerError)
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

func CreateUserInternal(db database.DBQuerier, email, firstName, lastName string) (models.User, error) {
	query := "INSERT INTO users (email, first_name, last_name) VALUES ($1, $2, $3) RETURNING user_id, email, first_name, last_name"
	var user models.User
	err := db.QueryRow(context.Background(), query, email, firstName, lastName).
		Scan(&user.UserID, &user.Email, &user.FirstName, &user.LastName)

	if err != nil {
		return models.User{}, err
	}

	log.Printf("✅ DEBUG: User created in DB: %+v", user)
	return user, nil
}
