package service

import (
	"net/http"
    "github.com/go-chi/chi/v5"
	"server/internal/handlers"
)


func NewRouter() http.Handler { 
	r := chi.NewRouter()

	r.Get("/health", handlers.HealthCheckHandler)

	return r

} 
