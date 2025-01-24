package routes

import (
	"context"
	"encoding/json"
	"megga-backend/models"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(router *mux.Router, db *pgxpool.Pool) {
	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		GetUsers(w, r, db)
	}).Methods("GET")
}

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
