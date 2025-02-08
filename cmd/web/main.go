package main

import (
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
	envFile := "env/.env.development"
	if os.Getenv("ENVIRONMENT") == "production" {
		envFile = "env/.env.production"
	}

	config.LoadAndValidateEnv(envFile)

	database.InitDB()
	defer database.CloseDB()

	cognitoConfig := middleware.CognitoConfig{
		UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		Region:     os.Getenv("AWS_REGION"),
	}

	frontendURL := os.Getenv("FRONTEND_URL")

	router := mux.NewRouter()

	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.WriteHeader(http.StatusOK)
	})

	router.Use(middleware.CORSConfig(frontendURL))
	router.Use(middleware.ValidateCognitoToken(cognitoConfig))

	routes.RegisterRoutes(router, database.DB)

	go func() {
		log.Println("🔄 Starting BLS Data Fetcher...")
		ticker := time.NewTicker(24 * time.Hour) // ✅ Fetch every 24 hours
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("🔄 Fetching latest BLS data...")
				blsData, err := services.FetchLatestBLSData()
				if err != nil {
					log.Printf("❌ Error fetching BLS data: %v", err)
					continue
				}
				err = services.SaveBLSData(database.DB, blsData)
				if err != nil {
					log.Printf("❌ Error saving BLS data: %v", err)
				}
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Starting server on :%s, allowing frontend URL: %s", port, frontendURL)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, router))
}
