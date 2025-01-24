package services

import (
	"github.com/gorilla/mux"
	"megga-backend/routes"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitRouter initializes the router and registers routes
func InitRouter(db *pgxpool.Pool) *mux.Router {
	router := mux.NewRouter()

	// Register user routes
	routes.RegisterUserRoutes(router, db)

	// Add other route registrations as needed
	return router
}
