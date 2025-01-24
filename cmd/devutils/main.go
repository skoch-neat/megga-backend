package main

import (
	"flag"
	"log"
	"megga-backend/devutils"
	"megga-backend/services"
)

func main() {
	services.LoadEnv()

	// Parse flags to determine action
	migrate := flag.Bool("migrate", false, "Run database migrations")
	seed := flag.Bool("seed", false, "Seed the database with test data")
	flag.Parse()

	// Initialize database connection
	services.InitDB()

	if *migrate {
		log.Println("Running migrations...")
		devutils.MigrateDB()
	}

	if *seed {
		log.Println("Seeding the database...")
		devutils.SeedDB()
	}
}
