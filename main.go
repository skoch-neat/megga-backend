package main

import (
	"log"
	"net/http"
	"os"

	"megga-backend/handlers"
	"megga-backend/middleware"
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

	// Cognito configuration
	cognitoConfig := middleware.CognitoConfig{
		UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		Region:     os.Getenv("AWS_REGION"),
	}

	// Set up the router
	router := mux.NewRouter()

	// Handle OPTIONS requests globally **before any middleware**
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.WriteHeader(http.StatusOK)
	})

	// Apply CORS middleware (before route registration)
	router.Use(middleware.CORSConfig(os.Getenv("FRONTEND_URL")))

	// Apply Cognito middleware globally to secure routes
	router.Use(middleware.ValidateCognitoToken(cognitoConfig))

	// Register user-related routes
	handlers.RegisterUserRoutes(router, database.DB)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
