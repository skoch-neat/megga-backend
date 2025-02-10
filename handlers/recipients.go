package handlers

import (
	"context"
	"encoding/json"
	"log"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func CreateRecipient(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var recipient models.Recipient

	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if recipient.Email == "" || recipient.FirstName == "" || recipient.LastName == "" || recipient.Designation == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO recipients (email, first_name, last_name, designation)
		VALUES ($1, $2, $3, $4)
		RETURNING recipient_id
	`
	err := db.QueryRow(context.Background(), query, recipient.Email, recipient.FirstName, recipient.LastName, recipient.Designation).
		Scan(&recipient.RecipientID)

	if err != nil {
		http.Error(w, "Database insert error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Recipient created successfully",
		"recipient": recipient,
	})
}

func GetRecipients(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var recipients []models.Recipient

	query := `
		SELECT recipient_id, email, first_name, last_name, designation
		FROM recipients
	`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var recipient models.Recipient
		if err := rows.Scan(
			&recipient.RecipientID, &recipient.Email, &recipient.FirstName, &recipient.LastName, &recipient.Designation,
		); err != nil {
			http.Error(w, "Error scanning recipients", http.StatusInternalServerError)
			return
		}
		recipients = append(recipients, recipient)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(recipients) == 0 {
		json.NewEncoder(w).Encode([]models.Recipient{})
	} else {
		json.NewEncoder(w).Encode(recipients)
	}
}

func GetRecipientByID(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		log.Println("❌ Invalid recipient ID received:", vars)
		http.Error(w, "Invalid recipient ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		log.Println("❌ Unable to convert ID:", idStr)
		http.Error(w, "Invalid recipient ID", http.StatusBadRequest)
		return
	}

	log.Printf("✅ Extracted Recipient ID: %d", id)

	var recipient models.Recipient
	query := `
		SELECT recipient_id, email, first_name, last_name, designation
		FROM recipients WHERE recipient_id = $1
	`
	err = db.QueryRow(context.Background(), query, id).Scan(
		&recipient.RecipientID, &recipient.Email, &recipient.FirstName, &recipient.LastName, &recipient.Designation,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Recipient not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipient)
}

func UpdateRecipient(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid recipient ID", http.StatusBadRequest)
		return
	}

	var recipient models.Recipient
	if err := json.NewDecoder(r.Body).Decode(&recipient); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE recipients
		SET email = $1, first_name = $2, last_name = $3, designation = $4
		WHERE recipient_id = $5
	`
	_, err = db.Exec(context.Background(), query, recipient.Email, recipient.FirstName, recipient.LastName, recipient.Designation, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Recipient updated successfully"})
}

func DeleteRecipient(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid recipient ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM recipients WHERE recipient_id = $1"
	res, err := db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Recipient not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Recipient deleted successfully"})
}

func RegisterRecipientRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/recipients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateRecipient(w, r, db)
		} else if r.Method == "GET" {
			GetRecipients(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "GET")

	router.HandleFunc("/recipients/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetRecipientByID(w, r, db)
		} else if r.Method == "PUT" {
			UpdateRecipient(w, r, db)
		} else if r.Method == "DELETE" {
			DeleteRecipient(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT", "DELETE")
}
