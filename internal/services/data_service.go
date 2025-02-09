package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"megga-backend/internal/database"
	"megga-backend/internal/models"

	"github.com/jackc/pgx/v4"
)

// ✅ Helper function to round a float to 2 decimal places
func roundFloat(val float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(val*multiplier) / multiplier
}

func SaveBLSData(db database.DBQuerier, blsData map[string]struct {
	Value  float64
	Year   string
	Period string
}) error {
	log.Println("💾 Saving BLS data to database...")

	if len(blsData) == 0 {
		log.Println("🔄 No BLS data to save.")
		return nil
	}

	// ✅ Define the UPDATE query at the start
	updateQuery := `UPDATE data 
					SET previous_value = latest_value, 
						latest_value = $1, 
						year = $2, 
						period = $3, 
						last_updated = NOW() 
					WHERE data_id = $4`

	// ✅ Start a single transaction for all updates
	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	// Ensure rollback if any failure occurs
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(context.Background()) // Ensure rollback on panic
			log.Printf("❌ Transaction panicked: %v", p)
		} else if err != nil {
			log.Println("🔄 Rolling back transaction due to error:", err)
			tx.Rollback(context.Background())
		}
	}()

	// ✅ Track whether any updates occurred
	updatesMade := false

	for seriesID, data := range blsData {
		var existing models.Data
		queryErr := db.QueryRow(context.Background(),
			"SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1",
			seriesID).Scan(&existing.DataID, &existing.LatestValue, &existing.PreviousValue, &existing.Year, &existing.Period)

		if queryErr == pgx.ErrNoRows {
			log.Printf("⚠️ No existing record found for series: %s, skipping.", seriesID)
			continue
		} else if queryErr != nil {
			err = fmt.Errorf("❌ Database query error: %w", queryErr)
			return err // Ensures rollback triggers via `defer`
		}

		// ✅ If the new data is NOT newer, **skip** updating
		if data.Year == existing.Year && data.Period == existing.Period {
			log.Printf("✅ No update needed for %s: %s-%s already exists. Skipping update.", seriesID, data.Year, data.Period)
			continue
		}

		// ✅ If an update occurs, set `updatesMade = true`
		updatesMade = true

		// ✅ Round the value before updating the database
		roundedValue := roundFloat(data.Value, 2)

		// ✅ Execute within the same transaction
		log.Printf("🔄 Updating %s with value: %.2f, Year: %s, Period: %s", seriesID, roundedValue, data.Year, data.Period)
		_, err = tx.Exec(context.Background(), updateQuery, roundedValue, data.Year, data.Period, existing.DataID)

		if err != nil {
			log.Printf("❌ Error executing update query for %s: %v", seriesID, err)
			return fmt.Errorf("error updating series %s: %w", seriesID, err)
		}

		log.Printf("✅ Successfully updated %s", seriesID)
	}

	// ✅ If NO updates were made, rollback instead of commit
	if !updatesMade {
		log.Println("🔄 No updates were made, rolling back transaction.")
		return tx.Rollback(context.Background()) // ✅ Return rollback instead of commit
	}

	// ✅ Commit transaction after all updates
	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Println("✅ All BLS data processed successfully.")
	return nil
}
