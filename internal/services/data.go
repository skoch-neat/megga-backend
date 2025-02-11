package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"megga-backend/internal/config"
	"megga-backend/internal/database"
	"megga-backend/internal/models"

	"github.com/jackc/pgx/v4"
)

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

	updateQuery := `UPDATE data 
					SET previous_value = latest_value, 
						latest_value = $1, 
						year = $2, 
						period = $3, 
						last_updated = NOW() 
					WHERE data_id = $4`

	insertQuery := `INSERT INTO data (name, series_id, unit, previous_value, latest_value, year, period, last_updated) 
					VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()) 
					RETURNING data_id`

	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(context.Background())
			log.Printf("❌ Transaction panicked: %v", p)
		} else if err != nil {
			log.Println("🔄 Rolling back transaction due to error:", err)
			tx.Rollback(context.Background())
		}
	}()

	changesMade := false

	for seriesID, data := range blsData {
		var existing models.Data
		queryErr := db.QueryRow(context.Background(),
			"SELECT data_id, latest_value, previous_value, year, period FROM data WHERE series_id = $1",
			seriesID).Scan(&existing.DataID, &existing.LatestValue, &existing.PreviousValue, &existing.Year, &existing.Period)

		if queryErr == pgx.ErrNoRows {
			log.Printf("⚠️ No existing record found for series: %s, inserting new record.", seriesID)

			info, exists := config.BLS_SERIES_INFO[seriesID]
			if !exists {
				log.Printf("❌ No series info found for %s, skipping.", seriesID)
				continue
			}

			roundedValue := roundFloat(data.Value, 2)

			_, err = tx.Exec(context.Background(), insertQuery, info.Name, seriesID, info.Unit, roundedValue, roundedValue, data.Year, data.Period)

			if err != nil {
				log.Printf("❌ Error inserting new record for %s: %v", seriesID, err)
				return fmt.Errorf("error inserting new record for series %s: %w", seriesID, err)
			}

			changesMade = true
			if config.IsDevelopmentMode() {
				log.Printf("✅ Successfully inserted new record for %s", seriesID)
			}
			continue

		} else if queryErr != nil {
			err = fmt.Errorf("❌ Database query error: %w", queryErr)
			return err
		}

		if data.Year == existing.Year && data.Period == existing.Period {
			if config.IsDevelopmentMode() {
				log.Printf("✅ No update needed for %s: %s-%s already exists. Skipping update.", seriesID, data.Year, data.Period)
			}
			continue
		}

		changesMade = true
		roundedValue := roundFloat(data.Value, 2)

		if config.IsDevelopmentMode() {
			log.Printf("🔄 Updating %s with value: %.2f, Year: %s, Period: %s", seriesID, roundedValue, data.Year, data.Period)
		}
		_, err = tx.Exec(context.Background(), updateQuery, roundedValue, data.Year, data.Period, existing.DataID)

		if err != nil {
			log.Printf("❌ Error executing update query for %s: %v", seriesID, err)
			return fmt.Errorf("error updating series %s: %w", seriesID, err)
		}

		if config.IsDevelopmentMode() {
			log.Printf("✅ Successfully updated %s", seriesID)
		}
	}

	if !changesMade {
		log.Println("🔄 No updates were made, rolling back transaction.")
		return tx.Rollback(context.Background())
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Println("✅ BLS data fetch complete.")
	return nil
}
