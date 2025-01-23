package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"megga-backend/routes"
	"megga-backend/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Ensure the database URL is set
	dsn := os.Getenv("DATABASE_URI")
	if dsn == "" {
		log.Fatal("DATABASE_URI is not set in the environment")
	}

	flag.Parse()

	// Initialize the database
	services.InitDB()

	// Initialize the router
	router := mux.NewRouter()

	// Register routes
	routes.RegisterUserRoutes(router)

	// Test endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to MEGGA!")
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("PORT") // Fallback to PORT environment variable
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
