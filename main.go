package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"megga-backend/services"
)

func main() {
	// Load environment variables
	services.LoadEnv()

	// Initialize the database
	services.InitDB()

	// Initialize the router
	router := services.InitRouter(services.DB)

	// Test endpoint - delete after demo
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to MEGGA!")
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal(("PORT environment variable is not set"))
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
