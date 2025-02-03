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

func CreateNotification(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var notification models.Notification

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if notification.UserID == 0 || notification.RecipientID == 0 || notification.ThresholdID == 0 || notification.UserMsg == "" || notification.RecipientMsg == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO notifications (user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg)
		VALUES ($1, $2, $3, NOW(), $4, $5)
		RETURNING notification_id
	`
	err := db.QueryRow(context.Background(), query, notification.UserID, notification.RecipientID, notification.ThresholdID, notification.UserMsg, notification.RecipientMsg).
		Scan(&notification.NotificationID)

	if err != nil {
		http.Error(w, "Database insert error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Notification created successfully",
		"notification": notification,
	})
}

func GetNotifications(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var notifications []models.Notification

	query := `
		SELECT notification_id, user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg
		FROM notifications
	`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var notification models.Notification
		if err := rows.Scan(
			&notification.NotificationID, &notification.UserID, &notification.RecipientID, &notification.ThresholdID,
			&notification.SentAt, &notification.UserMsg, &notification.RecipientMsg,
		); err != nil {
			http.Error(w, "Error scanning notifications", http.StatusInternalServerError)
			return
		}
		notifications = append(notifications, notification)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func GetNotificationByID(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	var notification models.Notification
	query := `
		SELECT notification_id, user_id, recipient_id, threshold_id, sent_at, user_msg, recipient_msg
		FROM notifications WHERE notification_id = $1
	`
	err = db.QueryRow(context.Background(), query, id).Scan(
		&notification.NotificationID, &notification.UserID, &notification.RecipientID,
		&notification.ThresholdID, &notification.SentAt, &notification.UserMsg, &notification.RecipientMsg,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}

func UpdateNotification(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	var notification models.Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE notifications
		SET user_msg = $1, recipient_msg = $2
		WHERE notification_id = $3
	`
	_, err = db.Exec(context.Background(), query, notification.UserMsg, notification.RecipientMsg, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification updated successfully"})
}

func DeleteNotification(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM notifications WHERE notification_id = $1"
	res, err := db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification deleted successfully"})
}

func RegisterNotificationRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateNotification(w, r, db)
		case http.MethodGet:
			GetNotifications(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods(http.MethodPost, http.MethodGet)

	router.HandleFunc("/notifications/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetNotificationByID(w, r, db)
		} else if r.Method == "PUT" {
			UpdateNotification(w, r, db)
		} else if r.Method == "DELETE" {
			DeleteNotification(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT", "DELETE")
}
