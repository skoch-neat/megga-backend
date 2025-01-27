package main

import (
	"flag"
	"log"
	"megga-backend/devutils"
	"megga-backend/services/database"
	"megga-backend/services/env"
)

func main() {
	env.LoadEnv()
	env.ValidateEnv()

	// Parse flags to determine action
	migrate := flag.Bool("migrate", false, "Run database migrations")
	seed := flag.Bool("seed", false, "Seed the database with test data")
	flag.Parse()

	// Ensure at least one flag is provided
	if !*migrate && !*seed {
		log.Println("No action specified. Use --migrate or --seed.")
		return
	}

	// Initialize the database connection
	db.InitDB()
	defer func() {
		// Close the database connection
		if db.DB != nil {
			db.DB.Close()
		}
	}()

	if *migrate {
		log.Println("Running migrations...")
		devutils.MigrateDB()
	}

	if *seed {
		log.Println("Seeding the database...")
		devutils.SeedDB()
	}
}
