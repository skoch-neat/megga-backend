package main

import (
	"flag"
	"log"
	"os"
	"megga-backend/internal/devutils"
	"megga-backend/internal/database"
	"megga-backend/internal/config"
)

func main() {
	envFile := "env/.env.development"
	if os.Getenv("ENVIRONMENT") == "production" {
		envFile = "env/.env.production"
	}

	log.Printf("🔍 Loading environment variables from: %s", envFile)
	config.LoadAndValidateEnv(envFile)

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
