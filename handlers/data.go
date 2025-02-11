package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"megga-backend/internal/config"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func CreateData(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	var data models.Data

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.SeriesID == "" {
		http.Error(w, "Missing required series_id", http.StatusBadRequest)
		return
	}

	if data.LatestValue < 0 {
		log.Printf("⚠️ [WARNING] Negative value received for %s: %.2f. Rejecting data.", data.SeriesID, data.LatestValue)
		http.Error(w, "Invalid data: latest_value cannot be negative", http.StatusBadRequest)
		return
	}

	if info, exists := config.BLS_SERIES_INFO[data.SeriesID]; exists {
		data.Name = info.Name
		data.Unit = info.Unit
	} else {
		http.Error(w, "Invalid series_id", http.StatusBadRequest)
		return
	}

	var existingID int
	checkQuery := `SELECT data_id FROM data WHERE series_id = $1`
	err := db.QueryRow(context.Background(), checkQuery, data.SeriesID).Scan(&existingID)

	if err == nil {
		http.Error(w, "Duplicate series_id: This series already exists.", http.StatusConflict)
		return
	} else if err != pgx.ErrNoRows {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}

	data.PreviousValue = data.LatestValue

	if data.Period == "" {
		data.Period = fmt.Sprintf("M%02d", time.Now().Month())
	}

	if data.Year == "" {
		data.Year = fmt.Sprintf("%d", time.Now().Year())
	}

	query := `
		INSERT INTO data (name, series_id, unit, previous_value, latest_value, last_updated, period, year)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6, $7)
		RETURNING data_id
	`
	err = db.QueryRow(r.Context(), query,
		data.Name, data.SeriesID, data.Unit, data.PreviousValue, data.LatestValue, data.Period, data.Year).
		Scan(&data.DataID)

	if err != nil {
		log.Printf("❌ Database error creating data: %v", err)
		http.Error(w, "Database insert error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Data created successfully",
		"data":    data,
	})
}

func GetData(w http.ResponseWriter, r *http.Request, db database.DBQuerier) {
	if config.IsDevelopmentMode() {
		log.Println("🔍 [DEBUG] Fetching all data...")
	}

	var dataEntries []models.Data
	query := `
		SELECT data_id, name, series_id, unit, previous_value, latest_value, last_updated, period, year
		FROM data
	`
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		if config.IsDevelopmentMode() {
			log.Printf("❌ [ERROR] Database query failed in GetData(): %v", err)
		}
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var data models.Data
		if err := rows.Scan(
			&data.DataID, &data.Name, &data.SeriesID, &data.Unit,
			&data.PreviousValue, &data.LatestValue, &data.LastUpdated, &data.Period, &data.Year,
		); err != nil {
			if config.IsDevelopmentMode() {
				log.Printf("❌ [ERROR] Error scanning data row in GetData(): %v", err)
			}
			http.Error(w, "Error scanning data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if config.IsDevelopmentMode() {
			log.Printf("✅ [DEBUG] Fetched Data Row: %+v", data)
		}
		dataEntries = append(dataEntries, data)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(dataEntries) == 0 {
		json.NewEncoder(w).Encode([]models.Data{})
	} else {
		json.NewEncoder(w).Encode(dataEntries)
	}
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
    SELECT data_id, name, series_id, unit, previous_value, latest_value, last_updated, period, year
    FROM data WHERE data_id = $1
`
	err = db.QueryRow(context.Background(), query, id).Scan(
		&data.DataID, &data.Name, &data.SeriesID, &data.Unit,
		&data.PreviousValue, &data.LatestValue, &data.LastUpdated, &data.Period, &data.Year,
	)

	if err == pgx.ErrNoRows {
		if config.IsDevelopmentMode() {
			log.Printf("✅ [DEBUG] No data found for ID: %d", id)
		}
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	} else if err != nil {
		if config.IsDevelopmentMode() {
			log.Printf("❌ [ERROR] Database error in GetDataByID(): %v", err)
		}
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

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
    SET previous_value = latest_value,
        latest_value = $1,
        name = $2,
        series_id = $3,
        unit = $4,
        period = $5,
        year = $6,
        last_updated = NOW()
    WHERE data_id = $7
`
	_, err = db.Exec(context.Background(), query, data.LatestValue, data.Name, data.SeriesID, data.Unit, data.Period, data.Year, id)

	if err != nil {
		if config.IsDevelopmentMode() {
			log.Printf("❌ [ERROR] Database error in UpdateData(): %v", err)
		}
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
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
		if config.IsDevelopmentMode() {
			log.Printf("❌ [ERROR] Failed to delete data in DeleteData(): %v", err)
		}
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if res.RowsAffected() == 0 {
		if config.IsDevelopmentMode() {
			log.Printf("✅ [DEBUG] No data found for deletion (ID: %d)", id)
		}
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
