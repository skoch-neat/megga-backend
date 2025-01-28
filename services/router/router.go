package router

import (
	"megga-backend/handlers"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

// InitRouter initializes the router
func InitRouter(dbPool *pgxpool.Pool) *mux.Router {
	router := mux.NewRouter()
	handlers.RegisterUserRoutes(router, dbPool)
	return router
}
