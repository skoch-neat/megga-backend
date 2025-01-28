package main

import (
	"log"
	"net/http"
	"os"

	"megga-backend/handlers"
	"megga-backend/services/database"
	"megga-backend/services/env"

	"github.com/gorilla/mux"
)

func main() {
	// Load and validate environment variables
	env.LoadEnv()
	env.ValidateEnv()

	// Initialize the database connection
	database.InitDB()
	defer database.CloseDB()

	// Set up the router
	router := mux.NewRouter()
	handlers.RegisterUserRoutes(router, database.DB)

	http.Handle("/", router)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
