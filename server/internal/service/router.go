package service

import (
	"net/http"
    "github.com/go-chi/chi/v5"
	"server/internal/handlers"
)


func NewRouter() http.Handler { 
	r := chi.NewRouter()
	r.Get("/health", handlers.HealthCheckHandler)
	
	r.Post("/api/v1/insert", handlers.InsertHandler)

	return r


}






