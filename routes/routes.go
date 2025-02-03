package routes

import (
	"megga-backend/handlers"
	"megga-backend/middleware"
	"megga-backend/services/database"
	"os"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router, db database.DBQuerier) {
	handlers.RegisterUserRoutes(router, db)
	handlers.RegisterThresholdRoutes(router, db)
	handlers.RegisterDataRoutes(router, db)
	handlers.RegisterNotificationRoutes(router, db)
	handlers.RegisterRecipientRoutes(router, db)
	handlers.RegisterThresholdRecipientRoutes(router, db)

	router.Use(middleware.ValidateCognitoToken(middleware.CognitoConfig{
		UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
		Region:    os.Getenv("AWS_REGION"),
	}))
}
