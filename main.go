package main

import (
	"log"
	"net/http"
	"os"

	"megga-backend/services/database"
	"megga-backend/services/env"
	"megga-backend/services/router"
)

func main() {
	// Load and validate environment variables
	env.LoadEnv()
	env.ValidateEnv()

	// Initialize the database and gracefully close it on exit
	db.InitDB()
	defer db.DB.Close()

	// Initialize the router
	router := router.InitRouter(db.DB)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
