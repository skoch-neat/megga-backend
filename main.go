package main

import (
	"log"
	"net/http"
	"os"

	"megga-backend/middleware"
	"megga-backend/routes"
	"megga-backend/services/database"
	"megga-backend/services/env"

	"github.com/gorilla/mux"
)

func main() {
	env.LoadEnv()
	env.ValidateEnv()

	database.InitDB()
	defer database.CloseDB()

	cognitoConfig := middleware.CognitoConfig{
		UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		Region:     os.Getenv("AWS_REGION"),
	}

	router := mux.NewRouter()

	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.WriteHeader(http.StatusOK)
	})

	router.Use(middleware.CORSConfig(os.Getenv("FRONTEND_URL")))

	router.Use(middleware.ValidateCognitoToken(cognitoConfig))

	routes.RegisterRoutes(router, database.DB)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
