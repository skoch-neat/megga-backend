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
	config.LoadAndValidateEnv()

	database.InitDB()
	defer database.CloseDB()

	// ✅ One-time initial fetch for BLS data
	log.Println("🛠️ Initializing BLS data...")
	err := services.FetchLatestBLSData(database.DB)
	if err != nil {
		log.Printf("❌ Error initializing BLS data: %v", err)
	} else {
		log.Println("✅ BLS data initialized successfully!")
	}

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

	// ✅ Periodic BLS Data Fetcher (delayed first execution)
	go func() {
		log.Println("⏳ Scheduling first BLS data fetch in 24 hours...")
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("🔄 Fetching latest BLS data...")
			err := services.FetchLatestBLSData(database.DB)
			if err != nil {
				log.Printf("❌ Error fetching BLS data: %v", err)
			} else {
				log.Println("✅ Successfully updated BLS data.")
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = "8080"

	log.Println("📌 Registered Routes:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err == nil {
			methods, _ := route.GetMethods()
			log.Printf("🔹 %s %s", methods, path)
		}
		return nil
	})

	log.Printf("🚀 Starting server on :%s, allowing frontend URL: %s", port, frontendURL)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, router))
}
