package services

import (
	"context"
	"fmt"
	"log"
	"megga-backend/internal/database"

	"github.com/jackc/pgx/v4"
)

// SaveBLSData stores the latest BLS data in the database
func SaveBLSData(db database.DBQuerier, blsData map[string]float64) error {
	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	for seriesID, value := range blsData {
		query := `
			UPDATE data
			SET latest_value = $1, last_updated = NOW()
			WHERE series_id = $2`
		_, err := tx.Exec(context.Background(), query, value, seriesID)
		if err != nil {
			log.Printf("❌ Error updating series %s: %v", seriesID, err)
			return fmt.Errorf("error updating series %s: %w", seriesID, err)
		}
		log.Printf("✅ Updated %s: %f", seriesID, value)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Println("✅ All BLS data updated successfully.")
	return nil
}
