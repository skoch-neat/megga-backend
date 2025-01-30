package routes

import (
	"megga-backend/handlers"
	"megga-backend/middleware"
	"megga-backend/services/database"

	"github.com/gorilla/mux"
)

// RegisterRoutes initializes all application routes
func RegisterRoutes(router *mux.Router, db database.DBQuerier) {
	handlers.RegisterUserRoutes(router, db)
	handlers.RegisterThresholdRoutes(router, db)
	handlers.RegisterDataRoutes(router, db)
	handlers.RegisterNotificationRoutes(router, db)
	handlers.RegisterRecipientRoutes(router, db)

	// Apply middleware if needed
	router.Use(middleware.ValidateCognitoToken(middleware.CognitoConfig{
		UserPoolID: "test-pool-id",
		Region:     "us-east-1",
	}))
}
