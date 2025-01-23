package routes

import (
	"context"
	"encoding/json"
	"megga-backend/models"
	"megga-backend/services"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(router *mux.Router) {
	router.HandleFunc("/users", GetUsers).Methods("GET")
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	// Query the database using services.DB
	rows, err := services.DB.Query(context.Background(), "SELECT user_id, email, first_name, last_name FROM users")
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
