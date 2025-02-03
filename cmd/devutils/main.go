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

	migrate := flag.Bool("migrate", false, "Run database migrations")
	seed := flag.Bool("seed", false, "Seed the database with test data")
	flag.Parse()

	if !*migrate && !*seed {
		log.Println("No action specified. Use --migrate or --seed.")
		return
	}

	database.InitDB()
	defer database.CloseDB()

	if *migrate {
		log.Println("Running migrations...")
		devutils.MigrateDB(database.DB)
	}

	if *seed {
		log.Println("Seeding the database...")
		devutils.SeedDB(database.DB)
	}
}
