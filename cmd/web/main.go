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

	initBLS := os.Getenv("INIT_BLS") == "true"

	go func() {
		if initBLS {
			log.Println("⏳ INIT_BLS set to true. Initializing BLS data...")
			err := services.FetchLatestBLSData(database.DB)
			if err != nil {
				log.Printf("❌ Error initializing BLS data: %v", err)
			} else {
				log.Println("✅ BLS data initialized successfully!")
			}
		} else {
			log.Println("⏳ INIT_BLS set to false. Skipping initial BLS data fetch.")
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

	if config.IsDevelopmentMode() {
		log.Printf("🚀 DEBUG: FRONTEND_URL from env = %s", frontendURL)
	}

	router := mux.NewRouter()

	router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.IsDevelopmentMode() {
			log.Println("✅ DEBUG: Handling global CORS preflight request (OPTIONS)")
		}
		frontendURL := os.Getenv("FRONTEND_URL")
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusNoContent)
	})

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
	}
	if config.IsDevelopmentMode() {
		log.Println("✅ PORT is set to: ", port)
	}

	addr := fmt.Sprintf(":%s", port)

	log.Printf("🚀 Starting server on port %s...", port)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
