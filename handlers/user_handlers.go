package handlers

import (
	"context"
	"encoding/json"
	"log"
	"megga-backend/models"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetUsers handles the retrieval of user data
func GetUsers(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	var users []models.User

	// Query the database
	rows, err := db.Query(context.Background(), "SELECT user_id, email, first_name, last_name FROM users")
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Populate users slice
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserID, &user.Email, &user.FirstName, &user.LastName)
		if err != nil {
			http.Error(w, "Error scanning user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	var user models.User

	// Decode the JSON request body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the user data
	if user.Email == "" || user.FirstName == "" || user.LastName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Insert the new user into the database
	_, err = db.Exec(context.Background(), `
		INSERT INTO users (email, first_name, last_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`, user.Email, user.FirstName, user.LastName)

	if err != nil {
		log.Printf("Error inserting user: %v\n", err)
		http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}
