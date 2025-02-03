package handlers

import (
	"context"
	"encoding/json"
	"megga-backend/models"
	"megga-backend/services/database"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func GetUsers(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var users []models.User

	rows, err := db.Query(context.Background(), "SELECT user_id, email, first_name, last_name FROM users")
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.UserID, &user.Email, &user.FirstName, &user.LastName); err != nil {
			http.Error(w, "Error scanning user data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUserByEmail(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	email := vars["email"]

	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	var user models.User
	query := "SELECT user_id, email, first_name, last_name FROM users WHERE email = $1"
	err := db.QueryRow(context.Background(), query, email).Scan(&user.UserID, &user.Email, &user.FirstName, &user.LastName)

	if err == pgx.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
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

func RegisterUserRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			CreateUser(w, r, db)
		case "GET":
			GetUsers(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "GET")

	router.HandleFunc("/users/email/{email}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetUserByEmail(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET")
}