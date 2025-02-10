package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"megga-backend/internal/config"
	"megga-backend/internal/database"
	"megga-backend/internal/middleware"
	"megga-backend/internal/routes"
	"megga-backend/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	config.LoadAndValidateEnv()
	database.InitDB()
	defer database.CloseDB()

	go func() {
		log.Println("⏳ Fetching BLS data now, then scheduling updates every 24 hours...")
		err := services.FetchLatestBLSData(database.DB)
		if err != nil {
			log.Printf("❌ Error initializing BLS data: %v", err)
		} else {
			log.Println("✅ BLS data initialized successfully!")
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			err := services.FetchLatestBLSData(database.DB)
			if err != nil {
				log.Printf("❌ Error fetching BLS data: %v", err)
			} else {
				log.Println("✅ Successfully updated BLS data.")
			}
		}
	}()

	cognitoConfig := middleware.CognitoConfig{
		UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		Region:     os.Getenv("AWS_REGION"),
	}
	frontendURL := os.Getenv("FRONTEND_URL")

	router := mux.NewRouter()

	router.Use(middleware.CORSConfig(frontendURL))
	router.Use(middleware.ValidateCognitoToken(cognitoConfig))
	routes.RegisterRoutes(router, database.DB)

	if config.IsDevelopmentMode() {
		log.Println("📌 Registered Routes:")
		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err == nil {
				methods, _ := route.GetMethods()
				log.Printf("🔹 %s %s", methods, path)
			}
			return nil
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("⚠️ PORT is not set, using default: ", port)
	} else {
		if config.IsDevelopmentMode() {
			log.Println("✅ PORT is set to: ", port)
		}
	}

	addr := fmt.Sprintf(":%s", port)

	log.Printf("🚀 Starting server on port %s...", port)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
