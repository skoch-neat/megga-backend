package routes

import (
	"megga-backend/handlers"
	"megga-backend/middleware"
	"megga-backend/services/database"
	"os"

	"github.com/gorilla/mux"
)

// RegisterRoutes initializes all application routes
func RegisterRoutes(router *mux.Router, db database.DBQuerier) {
	handlers.RegisterUserRoutes(router, db)
	handlers.RegisterThresholdRoutes(router, db)
	handlers.RegisterDataRoutes(router, db)
	handlers.RegisterNotificationRoutes(router, db)
	handlers.RegisterRecipientRoutes(router, db)
	handlers.RegisterThresholdRecipientRoutes(router, db)

	// Apply middleware if needed
	router.Use(middleware.ValidateCognitoToken(middleware.CognitoConfig{
		UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		Region:    os.Getenv("AWS_REGION"),
	}))
}
