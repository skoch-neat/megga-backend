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

// CreateData handles inserting a new Data entry
func CreateData(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var data models.Data

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if data.Name == "" || data.SeriesID == "" || data.Type == "" || data.Unit == "" || data.LatestValue == 0 || data.UpdateIntervalInDays == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// For a new data entry, PreviousValue should be initialized to the LatestValue
	data.PreviousValue = data.LatestValue

	// Insert query
	query := `
		INSERT INTO data (name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7)
		RETURNING data_id
	`

	err := db.QueryRow(context.Background(), query,
		data.Name, data.SeriesID, data.Type, data.Unit, data.PreviousValue, data.LatestValue, data.UpdateIntervalInDays).
		Scan(&data.DataID)

	if err != nil {
		http.Error(w, "Database insert error", http.StatusInternalServerError)
		return
	}

	// Response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Data created successfully",
		"data":    data,
	})
}

func GetData(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var dataEntries []models.Data

	query := `
		SELECT data_id, name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days
		FROM data
	`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var data models.Data
		if err := rows.Scan(
			&data.DataID, &data.Name, &data.SeriesID, &data.Type, &data.Unit,
			&data.PreviousValue, &data.LatestValue, &data.LastUpdated, &data.UpdateIntervalInDays,
		); err != nil {
			http.Error(w, "Error scanning data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		dataEntries = append(dataEntries, data)
	}

	// Return an empty array if no data exists
	if len(dataEntries) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.Data{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataEntries)
}

func GetDataByID(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
if err != nil || id <= 0 {
	http.Error(w, "Invalid data ID", http.StatusBadRequest)
	return
}


	var data models.Data
	query := `
		SELECT data_id, name, series_id, type, unit, previous_value, latest_value, last_updated, update_interval_in_days
		FROM data WHERE data_id = $1
	`
	err = db.QueryRow(context.Background(), query, id).Scan(
		&data.DataID, &data.Name, &data.SeriesID, &data.Type, &data.Unit,
		&data.PreviousValue, &data.LatestValue, &data.LastUpdated, &data.UpdateIntervalInDays,
	)

	if err == pgx.ErrNoRows {
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure proper JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func UpdateData(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid data ID", http.StatusBadRequest)
		return
	}

	var data models.Data
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE data
		SET name = $1, series_id = $2, type = $3, unit = $4, latest_value = $5, update_interval_in_days = $6
		WHERE data_id = $7
	`
	_, err = db.Exec(context.Background(), query, data.Name, data.SeriesID, data.Type, data.Unit, data.LatestValue, data.UpdateIntervalInDays, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data updated successfully"})
}

func DeleteData(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid data ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM data WHERE data_id = $1"
	res, err := db.Exec(context.Background(), query, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data deleted successfully"})
}

func RegisterDataRoutes(router *mux.Router, db database.DBQuerier) {
	router.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateData(w, r, db)
		} else if r.Method == "GET" {
			GetData(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("POST", "GET")

	router.HandleFunc("/data/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			GetDataByID(w, r, db)
		} else if r.Method == "PUT" {
			UpdateData(w, r, db)
		} else if r.Method == "DELETE" {
			DeleteData(w, r, db)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT", "DELETE")
}
